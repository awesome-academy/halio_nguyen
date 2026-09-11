// Removes .next WITHOUT following symlinks. `output: 'standalone'` writes
// directory symlinks under .next/standalone/node_modules that point at the
// real packages in node_modules/.pnpm; Next's own dist-dir clear-out on the
// next `dev`/`build` start can follow them on Windows and empty the real
// `next` package (observed repeatedly on Windows + pnpm). fs.rmSync unlinks
// a symlink instead of recursing into its target, so this is safe.
import { rmSync } from "node:fs";

rmSync(new URL("../.next", import.meta.url), { recursive: true, force: true });
