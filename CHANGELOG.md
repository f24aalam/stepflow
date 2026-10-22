# Changelog

## Unreleased — improvement proposal

- Propagate loading errors to callers.
- Guard empty and no-match list states and handle Unicode backspace.
- Make terminal steps accept Enter during dynamic flows.
- Add `Wizard.RunContext`, `LoadStep.RunContext`, and basic configuration validation.
- Add regression tests and cross-platform Go CI.
- Set normal terminal screen as default, preserving opt-in alternate screen.
- Add MIT license and refresh README and example.

Compatibility: default screen changed; list j/k shortcuts removed to support searching; context-aware loads are opt-in. Public import path remains `/pkg`.
