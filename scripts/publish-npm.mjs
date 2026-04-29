#!/usr/bin/env node
// Publish the LinkAI CLI to npm using the optionalDependencies pattern.
//
// Reads the GoReleaser dist/ output, stages a platform sub-package per
// (os, arch) pair, runs `npm publish` on each, then publishes the main
// `@linkai/cli` wrapper with optionalDependencies pinned to this version.
//
// Required env: RELEASE_VERSION (e.g. "v1.2.3") and NPM_TOKEN.
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

// (goOS, goArch) → (npm name, node process.platform, node process.arch, binary name)
const targets = [
  { goOS: "linux", goArch: "amd64", nodeOS: "linux", nodeArch: "x64", bin: "linkai" },
  { goOS: "linux", goArch: "arm64", nodeOS: "linux", nodeArch: "arm64", bin: "linkai" },
  { goOS: "darwin", goArch: "amd64", nodeOS: "darwin", nodeArch: "x64", bin: "linkai" },
  { goOS: "darwin", goArch: "arm64", nodeOS: "darwin", nodeArch: "arm64", bin: "linkai" },
  { goOS: "windows", goArch: "amd64", nodeOS: "win32", nodeArch: "x64", bin: "linkai.exe" },
  { goOS: "windows", goArch: "arm64", nodeOS: "win32", nodeArch: "arm64", bin: "linkai.exe" },
];

if (existsSync(stagingDir)) rmSync(stagingDir, { recursive: true, force: true });
mkdirSync(stagingDir, { recursive: true });

const subTemplateRaw = readFileSync(subTemplate, "utf8");

for (const t of targets) {
  const subName = `cli-${t.nodeOS}-${t.nodeArch}`;
  const fullName = `@linkai/${subName}`;
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

console.log(`==> publishing @linkai/cli@${version}`);
npmPublish(mainStage);

console.log("done");

function locateBinary({ goOS, goArch, bin }) {
  // GoReleaser v2 default layout: dist/{binary}_{os}_{arch}{,_v1,_v2,...}/{binary}
  const entries = readdirSync(distDir, { withFileTypes: true })
    .filter((e) => e.isDirectory())
    .map((e) => e.name)
    .filter((n) => n.startsWith(`linkai_${goOS}_${goArch}`));
  for (const dir of entries) {
    const candidate = join(distDir, dir, bin);
    if (existsSync(candidate)) return candidate;
  }
  fail(`could not find binary for ${goOS}_${goArch} under ${distDir}`);
}

function npmPublish(cwd) {
  const result = spawnSync("npm", ["publish", "--access", "public"], {
    cwd,
    stdio: "inherit",
    env: process.env,
  });
  if (result.status !== 0) {
    fail(`npm publish failed in ${cwd} (exit ${result.status})`);
  }
}

function fail(msg) {
  console.error(`publish-npm: ${msg}`);
  process.exit(1);
}
