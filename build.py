#!/usr/bin/env python3
"""
Build script for fyne-mc-world-manager (Go + Fyne GUI)
Compiles Windows x64 executable with proxy support.
"""

import argparse
import os
import shutil
import subprocess
import sys
import re
from pathlib import Path


def get_version() -> str:
    """Get version from main.go"""
    go_file = Path(__file__).parent / "main.go"
    content = go_file.read_text(encoding="utf-8")
    match = re.search(r'Version\s*=\s*"([^"]+)"', content)
    if match:
        return match.group(1)
    return "0.0.0"


def set_proxy(proxy: str):
    """Set proxy environment variables for Go"""
    if proxy:
        os.environ["HTTP_PROXY"] = proxy
        os.environ["HTTPS_PROXY"] = proxy
        os.environ["http_proxy"] = proxy
        os.environ["https_proxy"] = proxy
        os.environ["GOPROXY"] = "direct"
        os.environ["GONOPROXY"] = "none"
        print(f"Using proxy: {proxy}")


def run_go_mod_tidy():
    """Run go mod tidy to download dependencies"""
    print("\n[1/4] Running go mod tidy...")
    result = subprocess.run(
        ["go", "mod", "tidy"],
        capture_output=True,
        text=True,
        cwd=Path(__file__).parent,
    )
    if result.returncode != 0:
        print(f"Error: {result.stderr}")
        return False
    print("Dependencies downloaded successfully")
    return True


def build_executable(version: str, proxy: str):
    """Build Windows x64 executable"""
    print(f"\n[2/4] Building Windows x64 executable...")

    output_name = f"fyne-mc-world-manager-v{version}-windows-x64.exe"

    build_args = [
        "go",
        "build",
        "-ldflags",
        f"-s -w -X main.Version={version}",
        "-o",
        output_name,
    ]

    result = subprocess.run(
        build_args,
        capture_output=True,
        text=True,
        cwd=Path(__file__).parent,
    )

    if result.returncode != 0:
        print(f"Build error: {result.stderr}")
        return None

    print(f"Build successful: {output_name}")
    return output_name


def create_portable_zip(exe_name: str, version: str):
    """Create portable zip package"""
    print(f"\n[3/4] Creating portable zip...")

    zip_name = f"fyne-mc-world-manager-v{version}-windows-x64-green.zip"
    temp_dir = Path(__file__).parent / "temp_portable"
    temp_dir.mkdir(exist_ok=True)

    shutil.copy2(exe_name, temp_dir / exe_name)

    if shutil.make_archive(
        str(temp_dir),
        "zip",
        root_dir=temp_dir,
    ):
        final_zip = Path(__file__).parent / zip_name
        if final_zip.exists():
            final_zip.unlink()
        shutil.move(str(temp_dir) + ".zip", final_zip)
        print(f"Created: {zip_name}")

    shutil.rmtree(temp_dir)
    return zip_name


def cleanup():
    """Clean up build artifacts"""
    print(f"\n[4/4] Cleaning up...")
    for f in Path(__file__).parent.glob("*.exe"):
        if f.name != "fyne-mc-world-manager.exe":
            f.unlink()
    print("Cleanup complete")


def main():
    parser = argparse.ArgumentParser(
        description="Build fyne-mc-world-manager for Windows x64"
    )
    parser.add_argument(
        "--proxy",
        type=str,
        default="",
        help="HTTP/HTTPS proxy (e.g., http://192.168.31.233:7890)",
    )
    parser.add_argument(
        "--version",
        type=str,
        default="",
        help="Version string (auto-detected from main.go if not provided)",
    )
    args = parser.parse_args()

    version = args.version
    if not version:
        version = get_version()

    print(f"Building fyne-mc-world-manager v{version} for Windows x64")

    proxy = args.proxy
    if proxy:
        set_proxy(proxy)
    else:
        env_proxy = os.environ.get("HTTP_PROXY") or os.environ.get("HTTPS_PROXY")
        if env_proxy:
            print(f"Using proxy from environment: {env_proxy}")

    if not run_go_mod_tidy():
        sys.exit(1)

    exe_name = build_executable(version, proxy)
    if not exe_name:
        sys.exit(1)

    zip_name = create_portable_zip(exe_name, version)

    print("\n" + "=" * 50)
    print("Build completed successfully!")
    print(f"  EXE: {exe_name}")
    print(f"  ZIP: {zip_name}")
    print("=" * 50)


if __name__ == "__main__":
    main()
