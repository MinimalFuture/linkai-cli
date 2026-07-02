#!/usr/bin/env node
// Locate the platform-specific sub-package installed via optionalDependencies
// and forward argv + exit code to its native binary.
const { spawnSync } = require("node:child_process");
const path = require("node:path");

// npm rejects package names containing "win32" as spam, so the Windows
// sub-packages are named "windows" instead. Map the Node platform accordingly.
const platform = process.platform === "win32" ? "windows" : process.platform;
const pkg = `linkai-cli-${platform}-${process.arch}`;
const binName = process.platform === "win32" ? "linkai.exe" : "linkai";

let binary;
try {
  binary = require.resolve(`${pkg}/bin/${binName}`);
} catch (err) {
  console.error(
    `linkai: no prebuilt binary for ${process.platform}-${process.arch}.\n` +
      `Expected sub-package "${pkg}" to be installed via optionalDependencies.\n` +
      `If your install skipped optional deps, re-run with: npm i --include=optional`,
  );
  process.exit(1);
}

const result = spawnSync(binary, process.argv.slice(2), {
  stdio: "inherit",
  // Propagate the user's signals (Ctrl-C) to the child.
  windowsHide: false,
});

if (result.error) {
  console.error(result.error.message);
  process.exit(1);
}
process.exit(result.status ?? 0);
