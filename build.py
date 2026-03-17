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
from typing import Union


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


def run_go_mod_tidy(verbose: bool = False):
    """Run go mod tidy to download dependencies"""
    print("\n[1/4] Running go mod tidy...")
    env = os.environ.copy()
    if verbose:
        env["GOFLAGS"] = "-v"
    result = subprocess.run(
        ["go", "mod", "tidy"],
        cwd=Path(__file__).parent,
        env=env,
    )
    if result.returncode != 0:
        print(f"Error: go mod tidy failed with code {result.returncode}")
        return False
    print("Dependencies downloaded successfully")
    return True


import platform


def build_executable(version: str, proxy: str, verbose: bool = False):
    """Build Windows x64 executable"""
    print(f"\n[2/4] Building Windows x64 executable...")

    output_name = f"fyne-mc-world-manager-v{version}-windows-x64.exe"
    dist_dir = Path(__file__).parent / "dist"
    dist_dir.mkdir(exist_ok=True)
    output_path = dist_dir / output_name

    env = os.environ.copy()
    if verbose:
        env["GOFLAGS"] = "-v"
    if platform.system() == "Linux":
        # Cross-compiling from WSL/Linux
        env["GOOS"] = "windows"
        env["GOARCH"] = "amd64"
        env["CGO_ENABLED"] = "1"
        env["CC"] = "x86_64-w64-mingw32-gcc"
        print("Cross-compiling for Windows from Linux/WSL")
    else:
        # Native build on Windows
        env["CGO_ENABLED"] = "1"
        # Check gcc
        gcc_check = subprocess.run(
            ["gcc", "--version"],
            capture_output=True,
            text=True,
        )
        if gcc_check.returncode != 0:
            print("Error: C compiler 'gcc' not found.")
            print("Please install MinGW:")
            print("  - Download from https://www.mingw-w64.org/")
            print("  - Install to: D:\\SSoftwareFiles\\winget")
            print("  - Add to PATH: D:\\SSoftwareFiles\\winget\\bin")
            return None

    build_args = [
        "go",
        "build",
        "-ldflags",
        f"-s -w -X main.Version={version}",
        "-o",
        str(output_path),
    ]

    result = subprocess.run(
        build_args,
        capture_output=True,
        text=True,
        cwd=Path(__file__).parent,
        env=env,
    )

    if result.returncode != 0:
        print(f"Build error: {result.stderr}")
        return None

    print(f"Build successful: {output_name}")
    return output_path


def create_portable_zip(exe_name: Union[str, Path], version: str):
    """Create portable zip package"""
    print(f"\n[3/4] Creating portable zip...")

    zip_name = f"fyne-mc-world-manager-v{version}-windows-x64-green.zip"
    dist_dir = Path(__file__).parent / "dist"
    temp_dir = Path(__file__).parent / "temp_portable"
    temp_dir.mkdir(exist_ok=True)

    shutil.copy2(exe_name, temp_dir / exe_name)

    if shutil.make_archive(
        str(temp_dir),
        "zip",
        root_dir=temp_dir,
    ):
        final_zip = dist_dir / zip_name
        if final_zip.exists():
            final_zip.unlink()
        shutil.move(str(temp_dir) + ".zip", final_zip)
        print(f"Created: {zip_name}")

    shutil.rmtree(temp_dir)
    return zip_name


def cleanup():
    """Clean up build artifacts"""
    print(f"\n[4/4] Cleaning up...")
    dist_dir = Path(__file__).parent / "dist"
    if dist_dir.exists():
        shutil.rmtree(dist_dir)
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
    parser.add_argument(
        "--verbose",
        "-v",
        action="store_true",
        help="Enable verbose output for Go commands",
    )
    args = parser.parse_args()

    version = args.version
    if not version:
        version = get_version()

    verbose = args.verbose

    print(f"Building fyne-mc-world-manager v{version} for Windows x64")

    proxy = args.proxy
    if proxy:
        set_proxy(proxy)
    else:
        env_proxy = os.environ.get("HTTP_PROXY") or os.environ.get("HTTPS_PROXY")
        if env_proxy:
            print(f"Using proxy from environment: {env_proxy}")

    if not run_go_mod_tidy(verbose):
        sys.exit(1)

    exe_name = build_executable(version, proxy, verbose)
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
