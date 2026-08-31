#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
APP_DIR="$ROOT_DIR/build/native/macos/SkillFlow.app"
SWIFT_PACKAGE_DIR="$ROOT_DIR/native/macos/SkillFlow"
SWIFT_BINARY="$SWIFT_PACKAGE_DIR/.build/release/SkillFlow"

echo "native macos build started: target=$APP_DIR"
rm -rf "$ROOT_DIR/build/native/macos"
mkdir -p "$APP_DIR/Contents/MacOS" "$APP_DIR/Contents/Resources"

swift build --package-path "$SWIFT_PACKAGE_DIR" -c release

# Ensure frontend/dist exists for the //go:embed all:frontend/dist directive.
# The native daemon runs with --daemon-only and never serves the Wails UI,
# but the embed directive is evaluated at compile time.
FRONTEND_DIST="$ROOT_DIR/cmd/skillflow/frontend/dist"
if [[ ! -d "$FRONTEND_DIST" ]]; then
  mkdir -p "$FRONTEND_DIST"
  echo '<!DOCTYPE html><html><head><title>SkillFlow</title></head><body></body></html>' > "$FRONTEND_DIST/index.html"
fi

go build -trimpath -ldflags "-s -w" -o "$APP_DIR/Contents/MacOS/skillflowd" "$ROOT_DIR/cmd/skillflow"
cp "$SWIFT_BINARY" "$APP_DIR/Contents/MacOS/SkillFlow"

if [[ -f "$ROOT_DIR/cmd/skillflow/build/darwin/iconfile.icns" ]]; then
  cp "$ROOT_DIR/cmd/skillflow/build/darwin/iconfile.icns" "$APP_DIR/Contents/Resources/iconfile.icns"
fi

cat > "$APP_DIR/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleDevelopmentRegion</key><string>en</string>
  <key>CFBundleDisplayName</key><string>SkillFlow</string>
  <key>CFBundleExecutable</key><string>SkillFlow</string>
  <key>CFBundleIdentifier</key><string>com.shinerio.skillflow.native</string>
  <key>CFBundleInfoDictionaryVersion</key><string>6.0</string>
  <key>CFBundleName</key><string>SkillFlow</string>
  <key>CFBundlePackageType</key><string>APPL</string>
  <key>CFBundleShortVersionString</key><string>1.0</string>
  <key>CFBundleVersion</key><string>1</string>
  <key>LSMinimumSystemVersion</key><string>13.0</string>
  <key>NSHighResolutionCapable</key><true/>
  <key>NSPrincipalClass</key><string>NSApplication</string>
</dict>
</plist>
PLIST
printf 'APPL????' > "$APP_DIR/Contents/PkgInfo"
chmod +x "$APP_DIR/Contents/MacOS/SkillFlow" "$APP_DIR/Contents/MacOS/skillflowd"

if command -v codesign >/dev/null 2>&1; then
  codesign --force --deep --sign - "$APP_DIR"
fi

echo "native macos build completed: target=$APP_DIR"
