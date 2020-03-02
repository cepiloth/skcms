#!/usr/bin/python2.7

import os
import subprocess
import sys

ninja = sys.argv[1]

def call(cmd):
  subprocess.check_call(cmd, shell=True)

def append(path, line):
  with open(path, 'a') as f:
    print >>f, line

print "Hello from {platform} in {cwd}!".format(platform=sys.platform,
                                               cwd=os.getcwd())

if 'darwin' in sys.platform:
  # Get Xcode from CIPD using mac_toolchain tool.
  mac_toolchain = os.path.join(os.getcwd(), sys.argv[3])
  xcode_app_path = os.path.join(os.getcwd(), sys.argv[4])
  # See mapping of Xcode version to Xcode build version here:
  # https://chromium.googlesource.com/chromium/tools/build/+/master/scripts/slave/recipe_modules/ios/api.py#37
  XCODE_BUILD_VERSION = '9c40b'
  call(('{mac_toolchain}/mac_toolchain install '
        '-kind mac '
        '-xcode-version {xcode_build_version} '
        '-output-dir {xcode_app_path}').format(
            mac_toolchain=mac_toolchain,
            xcode_build_version=XCODE_BUILD_VERSION,
            xcode_app_path=xcode_app_path))
  call('sudo xcode-select -switch {xcode_app_path}'.format(
      xcode_app_path=xcode_app_path))

  # Our Mac bot toolchains are too old for LSAN.
  append('skcms/build/clang.lsan', 'disabled = true')

  call('{ninja}/ninja -C skcms -k 0'.format(ninja=ninja))

elif 'linux' in sys.platform:
  # Point to clang in our clang_linux package.
  clang_linux = os.path.realpath(sys.argv[3])
  append('skcms/build/clang', 'cc  = {}/bin/clang  '.format(clang_linux))
  append('skcms/build/clang', 'cxx = {}/bin/clang++'.format(clang_linux))

  call('{ninja}/ninja -C skcms -k 0'.format(ninja=ninja))

else:  # Windows
  win_toolchain = os.path.realpath(sys.argv[2])
  os.environ['PATH'] = win_toolchain + '\\VC\\Tools\\MSVC\\14.16.27023\\bin\\HostX64\\x64;' + os.environ['PATH']
  os.environ['INCLUDE'] = win_toolchain + '\\VC\\Tools\\MSVC\\14.16.27023\\include;'
  os.environ['INCLUDE'] += win_toolchain + '\\win_sdk\\Include\\10.0.17763.0\\shared;'
  os.environ['INCLUDE'] += win_toolchain + '\\win_sdk\\Include\\10.0.17763.0\\ucrt;'
  os.environ['INCLUDE'] += win_toolchain + '\\win_sdk\\Include\\10.0.17763.0\\um;'
  os.environ['LIB'] = win_toolchain + '\\VC\\Tools\\MSVC\\14.16.27023\\lib\\x64;'
  os.environ['LIB'] += win_toolchain + '\\win_sdk\\Lib\\10.0.17763.0\\um\\x64;'
  os.environ['LIB'] += win_toolchain + '\\win_sdk\\Lib\\10.0.17763.0\\ucrt\\x64;'

  call('{ninja}\\ninja.exe -C skcms -f msvs.ninja -k 0'.format(ninja=ninja))
