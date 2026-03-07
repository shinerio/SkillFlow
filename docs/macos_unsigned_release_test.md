# macOS Unsigned Release Test Checklist

This checklist is for SkillFlow's unsigned macOS release flow (Ad-hoc signed app + DMG, no Developer ID / notarization).

## Automated CI gates (tag release)

- Build `darwin/amd64` and `darwin/arm64` artifacts.
- Sign `.app` and `.dmg` with ad-hoc identity.
- Verify app signature with:
  - `codesign --verify --deep --strict --verbose=2`
  - `lipo -archs` architecture check (`x86_64` or `arm64`)
- Mount DMG and re-run:
  - app presence check
  - `codesign --verify --deep --strict --verbose=2`
  - `spctl --assess --type execute --verbose=4` output sanity check (must not include `damaged`)
- Compare release-asset SHA256 against build-stage SHA256.

## Manual acceptance (required per release)

Run on two machines:

- Apple Silicon macOS
- Intel macOS

Steps:

1. Download the DMG from GitHub Release using a browser (keep quarantine attribute).
2. Open DMG and drag `SkillFlow.app` to `/Applications`.
3. First launch should be blocked by Gatekeeper (expected for unsigned distribution).
4. Open `System Settings > Privacy & Security`.
5. Confirm `Open Anyway` appears for SkillFlow, then launch succeeds.
6. Confirm no `SkillFlow is damaged and can’t be opened` message appears.

## Negative test (weekly or pre-release)

1. Tamper with app bundle contents after signing.
2. Re-run `codesign --verify --deep --strict --verbose=2`.
3. Expect verification failure.

Optional sanity check:

1. Build a temporary branch with ad-hoc signing step disabled.
2. Confirm artifact quality regresses (to validate CI gate value).
