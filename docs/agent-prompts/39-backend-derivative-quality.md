# Prompt 39 — Backend derivative quality improvement

Replace nearest-neighbor scaling with higher-quality image resampling.

Implement:
- add `golang.org/x/image` dependency
- replace manual pixel copy in `derive.go` with `draw.CatmullRom` (or `draw.BiLinear`)
- preserve existing cache layout and file naming
- update tests to verify output dimensions are correct

Do not change the cache structure or API behavior.
