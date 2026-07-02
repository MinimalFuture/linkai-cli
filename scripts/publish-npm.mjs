#!/usr/bin/env node
// Publish the LinkAI CLI to npm using the optionalDependencies pattern.
//
// Reads the GoReleaser dist/ output, stages a platform sub-package per
// (os, arch) pair, runs `npm publish` on each, then publishes the main
// `linkai-cli` wrapper with optionalDependencies pinned to this version.
//
// Auth: in CI this relies on npm Trusted Publishing (OIDC) — no NPM_TOKEN.
// The Release workflow grants `id-token: write` and upgrades npm to >= 11.5.1;
// every package (main + each platform sub-package) must have a Trusted Publisher
// configured on npmjs.com bound to this repo + workflow file (release.yml).
// Under OIDC, npm attaches build provenance automatically.
//
// Required env: RELEASE_VERSION (e.g. "v1.2.3").
// Optional env: NPM_NO_PROVENANCE=1 to disable provenance (e.g. for a local
//   token-based publish where OIDC is unavailable).
// Run from the repo root after `goreleaser release` has populated dist/.

import { spawnSync } from "node:child_process";
import { existsSync, mkdirSync, readdirSync, readFileSync, rmSync, writeFileSync, chmodSync, copyFileSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const repoRoot = resolve(__dirname, "..");
const distDir = resolve(repoRoot, "dist");
const stagingDir = resolve(repoRoot, "dist", "npm-staging");
const mainPkgDir = resolve(repoRoot, "npm", "cli");
const subTemplate = resolve(repoRoot, "npm", "cli-template", "package.json");

const rawVersion = process.env.RELEASE_VERSION || "";
if (!rawVersion) {
  fail("RELEASE_VERSION is required (e.g. v1.2.3)");
}
const version = rawVersion.replace(/^v/, "");

// (goOS, goArch) → npm sub-package metadata.
//   pkgOS   segment used in the package NAME. Windows uses "windows" (not
//           "win32") because npm rejects "win32" in package names as spam.
//   nodeOS  value for the package.json "os" field — must be the Node
//           process.platform value ("win32"), so npm installs it on Windows.
//   nodeArch value for the "cpu" field and the wrapper's process.arch lookup.
const targets = [
  { goOS: "linux", goArch: "amd64", pkgOS: "linux", nodeOS: "linux", nodeArch: "x64", bin: "linkai" },
  { goOS: "linux", goArch: "arm64", pkgOS: "linux", nodeOS: "linux", nodeArch: "arm64", bin: "linkai" },
  { goOS: "darwin", goArch: "amd64", pkgOS: "darwin", nodeOS: "darwin", nodeArch: "x64", bin: "linkai" },
  { goOS: "darwin", goArch: "arm64", pkgOS: "darwin", nodeOS: "darwin", nodeArch: "arm64", bin: "linkai" },
  { goOS: "windows", goArch: "amd64", pkgOS: "windows", nodeOS: "win32", nodeArch: "x64", bin: "linkai.exe" },
  { goOS: "windows", goArch: "arm64", pkgOS: "windows", nodeOS: "win32", nodeArch: "arm64", bin: "linkai.exe" },
];

if (existsSync(stagingDir)) rmSync(stagingDir, { recursive: true, force: true });
mkdirSync(stagingDir, { recursive: true });

const subTemplateRaw = readFileSync(subTemplate, "utf8");

for (const t of targets) {
  const fullName = `linkai-cli-${t.pkgOS}-${t.nodeArch}`;
  const subName = fullName;
  const stageRoot = join(stagingDir, subName);
  mkdirSync(join(stageRoot, "bin"), { recursive: true });

  const binarySrc = locateBinary(t);
  copyFileSync(binarySrc, join(stageRoot, "bin", t.bin));
  if (t.goOS !== "windows") chmodSync(join(stageRoot, "bin", t.bin), 0o755);

  const pkgJson = subTemplateRaw
    .replace(/__NAME__/g, fullName)
    .replace(/__VERSION__/g, version)
    .replace(/__OS__/g, t.goOS)
    .replace(/__ARCH__/g, t.goArch)
    .replace(/__NODE_OS__/g, t.nodeOS)
    .replace(/__NODE_ARCH__/g, t.nodeArch);
  writeFileSync(join(stageRoot, "package.json"), pkgJson);

  console.log(`==> publishing ${fullName}@${version}`);
  npmPublish(stageRoot);
}

// Stage the main wrapper with the resolved version.
const mainStage = join(stagingDir, "cli-main");
mkdirSync(mainStage, { recursive: true });
mkdirSync(join(mainStage, "bin"), { recursive: true });
copyFileSync(join(mainPkgDir, "bin", "linkai.js"), join(mainStage, "bin", "linkai.js"));
chmodSync(join(mainStage, "bin", "linkai.js"), 0o755);

const mainPkg = JSON.parse(readFileSync(join(mainPkgDir, "package.json"), "utf8"));
mainPkg.version = version;
mainPkg.optionalDependencies = Object.fromEntries(
  Object.keys(mainPkg.optionalDependencies || {}).map((k) => [k, version]),
);
writeFileSync(join(mainStage, "package.json"), JSON.stringify(mainPkg, null, 2) + "\n");

console.log(`==> publishing linkai-cli@${version}`);
npmPublish(mainStage);

console.log("done");

function locateBinary({ goOS, goArch, bin }) {
  // GoReleaser v2 build dirs are named "{id}_{os}_{arch}{,_v1,_v8.0,...}", where
  // {id} defaults to the project name (linkai-cli), not the binary name. Match
  // any dir ending in "_{os}_{arch}[_variant]" and containing the binary, so we
  // don't depend on the exact id or GOAMD64/GOARM variant suffix.
  const re = new RegExp(`_${goOS}_${goArch}(_[^/]+)?$`);
  const entries = readdirSync(distDir, { withFileTypes: true })
    .filter((e) => e.isDirectory())
    .map((e) => e.name)
    .filter((n) => re.test(n));
  for (const dir of entries) {
    const candidate = join(distDir, dir, bin);
    if (existsSync(candidate)) return candidate;
  }
  fail(`could not find binary for ${goOS}_${goArch} under ${distDir}`);
}

function npmPublish(cwd) {
  const args = ["publish", "--access", "public"];
  // Provenance is generated automatically under OIDC trusted publishing.
  // Pass it explicitly so the run fails loudly if the OIDC context is missing,
  // unless a local token-based publish opts out via NPM_NO_PROVENANCE.
  if (process.env.NPM_NO_PROVENANCE !== "1") {
    args.push("--provenance");
  }
  // Capture output (not inherit) so we can treat an already-published version
  // as success, making re-runs idempotent when an earlier attempt published
  // some packages before failing on a later one.
  const result = spawnSync("npm", args, { cwd, encoding: "utf8", env: process.env });
  const out = `${result.stdout || ""}${result.stderr || ""}`;
  process.stdout.write(out);
  if (result.status === 0) return;
  if (/EPUBLISHCONFLICT|cannot publish over|previously published versions/i.test(out)) {
    console.log(`==> already published, skipping (${cwd})`);
    return;
  }
  fail(`npm publish failed in ${cwd} (exit ${result.status})`);
}

function fail(msg) {
  console.error(`publish-npm: ${msg}`);
  process.exit(1);
}
