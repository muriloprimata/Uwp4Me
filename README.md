<div align="center">

# 🎮 UWP4Me

**Build, package and sign Xbox UWP apps straight from your Android phone.**
No PC. No Visual Studio. Just Termux.

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
![Platform](https://img.shields.io/badge/platform-Termux%20(Android)-3DDC84?logo=android&logoColor=white)
![Target](https://img.shields.io/badge/target-Xbox%20Dev%20Mode-107C10?logo=xbox&logoColor=white)
![Made with Go](https://img.shields.io/badge/made%20with-Go-00ADD8?logo=go&logoColor=white)
![Status](https://img.shields.io/badge/status-alpha-orange)

`#xbox` `#uwp` `#termux` `#android` `#golang` `#msix` `#cpp-winrt` `#homebrew` `#dev-mode` `#clang` `#cross-compilation` `#no-pc`

</div>

---

## ✨ What is UWP4Me?

Making apps and games for an Xbox in **Dev Mode** normally means owning a Windows PC with Visual Studio.
**UWP4Me removes that requirement.** It is a command-line toolkit that turns an Android phone running
[Termux](https://termux.dev) into a complete Xbox app workshop:

| Step | What happens | Done with |
| --- | --- | --- |
| 🧰 **Compile** | C++/WinRT code becomes a Windows x64 UWP executable | `clang` cross-compiling to `x86_64-pc-windows-msvc` + `lld-link` |
| 📚 **SDK** | The Windows SDK and C runtime are downloaded **from Microsoft's own servers** | UWP4Me's built-in downloader (written in Go) |
| 📦 **Package** | The app folder becomes a `.msix` package | Microsoft's [MSIX Packaging SDK](https://github.com/microsoft/msix-packaging), built on your phone |
| ✍️ **Sign** | The package gets signed with your own certificate | `osslsigncode` + `openssl` |
| 🚀 **Install** | The signed package is installed on the console | Xbox **Device Portal** (browser) |

> It was born from a simple rule: **everything must run on the phone.** No PC, no CI, no "just use GitHub Actions".

### ✅ Proven on real hardware

A hello-world app and a complete **Snake game** (Direct3D 11 + gamepad) were compiled, packaged and signed
entirely inside a 32-bit Termux on an Android phone, then installed and played on an **Xbox Series S** in Dev Mode.

---

## 🚧 Status: alpha

UWP4Me is young. The foundation is solid and tested; the one-command workflow is still being built.

| Feature | State |
| --- | --- |
| `setup` — installs tools, builds `makemsix` with Termux patches | ✅ Working |
| `sdk` — license prompt, SDK/CRT download, extraction, case fixing | ✅ Working |
| `doctor` — checks the whole toolchain, including a real test build | ✅ Working |
| `clean` — removes built tools, SDK and caches | ✅ Working |
| `init` — project templates (hello, snake) | 🚧 Coming soon |
| `build` — compile + link in one command | 🚧 Coming soon (manual script in the tutorial below) |
| `pack` — package + sign in one command | 🚧 Coming soon (manual script in the tutorial below) |
| `deploy` — install straight on the Xbox from the terminal | 🚧 Coming soon (today: Device Portal in the browser) |
| Slimmer SDK (only what UWP needs) | 🗺️ Planned |
| Pinned SDK versions for reproducible builds | 🗺️ Planned |
| Prebuilt release binaries | 🗺️ Planned |

---

## 📋 Requirements

- 📱 An **Android phone** with **[Termux](https://github.com/termux/termux-app/releases)** (install from F-Droid or GitHub, not the Play Store). 32-bit and 64-bit ARM are both fine.
- 💾 About **2 GB of free space** (SDK cache plus extracted files; more while downloading).
- 📶 Wi-Fi for the first setup (hundreds of MB are downloaded).
- 🎮 An **Xbox One or Series S|X in Developer Mode**, on the **same network** as the phone.
- ⏱️ Patience on the first run: building `makemsix` on a phone takes **10 to 15 minutes**.

> 🔋 Plug the phone in and keep Termux in the foreground. Run `termux-wake-lock` so Android does not put it to sleep.

---

## 🚀 Quick start

```bash
# 1. Install Go and git inside Termux
pkg update && pkg install -y git golang

# 2. Get UWP4Me
git clone https://github.com/muriloprimata/uwp4me.git
cd uwp4me
go build -o uwp4me ./cmd/uwp4me

# 3. Put it on your PATH (optional)
cp uwp4me $PREFIX/bin/

# 4. Run it. The first run welcomes you and sets everything up.
uwp4me
```

Or install straight with Go:

```bash
go install github.com/muriloprimata/uwp4me/cmd/uwp4me@latest
export PATH="$PATH:$HOME/go/bin"
```

Then download the Windows SDK and check that everything works:

```bash
uwp4me sdk       # asks you to accept Microsoft's license, then downloads the SDK
uwp4me doctor    # should end with "0 problems"
```

---

## 🧭 Commands

| Command | What it does |
| --- | --- |
| `uwp4me setup` | Installs the needed Termux packages and builds `makemsix`. Runs on its own the first time. |
| `uwp4me sdk` | Asks for Microsoft's license once, downloads the SDK and C runtime, extracts them and builds the ready-to-use layout. |
| `uwp4me sdk plan` | Only lists the newest packages found in Microsoft's catalog. Downloads nothing. |
| `uwp4me sdk layout` | Rebuilds the SDK layout from files already downloaded (no network). |
| `uwp4me doctor` | Checks tools, `makemsix`, license and SDK, then compiles and links a real test program. |
| `uwp4me clean` | Removes the built tools (`makemsix` and its sources). Asks before deleting. |
| `uwp4me clean --sdk` | Also removes the SDK and the download cache. |
| `uwp4me clean --all` | Removes everything in `~/.uwp4me`, including the accepted license. |
| `uwp4me version` | Prints the version. |

**Options and environment variables**

| Name | Effect |
| --- | --- |
| `--accept-license` / `UWP4ME_ACCEPT_LICENSE=1` | Accepts Microsoft's terms without asking (for scripts). |
| `--yes` (with `clean`) | Does not ask for confirmation. |
| `UWP4ME_REBUILD=1` | Forces `setup` to rebuild `makemsix` from scratch. |
| `UWP4ME_REFRESH=1` | Downloads Microsoft's package catalog again instead of using the cache. |
| `NO_COLOR=1` | Disables colors in the output. |

---

## 🧠 How it works

```
 your C++/WinRT code
        │
        ▼
 clang++ --target=x86_64-pc-windows-msvc      (cross compiler running on Android)
        │      uses headers and libs from ──►  Windows SDK + MSVC runtime
        ▼                                       downloaded by UWP4Me from Microsoft
 lld-link /appcontainer                        (links a UWP executable)
        │
        ▼
 makemsix pack                                 (Microsoft MSIX Packaging SDK, built on your phone)
        │
        ▼
 osslsigncode sign                             (your own self-signed certificate)
        │
        ▼
 Xbox Device Portal  ──►  app installed on the console 🎉
```

### Why a custom SDK downloader?

The usual helper tools are not available for 32-bit ARM Termux, so UWP4Me reads Microsoft's Visual Studio
catalog itself, downloads only the packages a UWP app needs, verifies each one with its **SHA-256**, resumes
interrupted downloads, and extracts the installers with `msitools`.

### Why all those symbolic links?

Windows ignores upper and lower case in file names. **Android does not.** The SDK headers ask for
`DriverSpecs.h` while the file is called `driverspecs.h`, and so on, thousands of times.
UWP4Me reads every `#include` in the headers and creates links with the exact spelling, so the compiler finds
everything.

### What gets patched?

`makemsix` does not build on Android out of the box. UWP4Me downloads the MSIX Packaging SDK **at a pinned
commit** and applies a small patch (included in this repo, `patches/msix-termux.patch`) that:

- builds the XML parser with C++17, as the Termux ICU headers require;
- avoids a name clash with Android's `LOCK_WRITE` macro;
- provides the OpenSSL `async_local_cleanup` function that bionic needs.

### Where do files go?

```
~/.uwp4me/
├── bin/makemsix            the packager
├── lib/libmsix.so          its library
├── src/msix-packaging/     pinned Microsoft sources (patched)
├── cache/                  downloads, extracted SDK, Microsoft catalog
├── sdk/                    ready-to-use headers and libs (symbolic links)
├── license-accepted        your acceptance of Microsoft's terms
└── msix-build.log          full log of the makemsix build
```

---

## 📖 Tutorial: your first Xbox app, from the phone

### Part 1 — Prepare the Xbox

1. On the Xbox, enable **Developer Mode** (install the *Dev Mode Activation* app and follow its steps) and reboot into Dev Home.
2. In **Dev Home**, open the **remote access settings** and **enable Device Portal**. Create a **username and password**.
3. Write down the **IP address** the console shows.
4. Make sure the phone is on the **same Wi-Fi network**.

### Part 2 — Prepare the phone

```bash
termux-setup-storage     # lets Termux copy files to the Download folder
termux-wake-lock         # keeps Termux alive during long steps
uwp4me                   # first run: installs tools and builds makemsix (10-15 min)
uwp4me sdk               # accept the license and download the SDK
uwp4me doctor            # expect: 0 problems
```

### Part 3 — Write the app

Create a folder and a minimal C++/WinRT app (a window that stays open):

```bash
mkdir -p ~/hello && cd ~/hello
```

<details>
<summary><b>hello.cpp</b> (click to expand)</summary>

```bash
cat > hello.cpp <<'EOF'
#include <winrt/Windows.Foundation.h>
#include <winrt/Windows.ApplicationModel.Core.h>
#include <winrt/Windows.UI.Core.h>

using namespace winrt;
using namespace winrt::Windows::ApplicationModel::Core;
using namespace winrt::Windows::UI::Core;

struct App : implements<App, IFrameworkViewSource, IFrameworkView>
{
    IFrameworkView CreateView() { return *this; }
    void Initialize(CoreApplicationView const&) {}
    void Load(hstring const&) {}
    void Uninitialize() {}
    void SetWindow(CoreWindow const&) {}
    void Run()
    {
        CoreWindow window = CoreWindow::GetForCurrentThread();
        window.Activate();
        window.Dispatcher().ProcessEvents(CoreProcessEventsOption::ProcessUntilQuit);
    }
};

int __stdcall wWinMain(void*, void*, wchar_t*, int)
{
    CoreApplication::Run(make<App>());
    return 0;
}
EOF
```

</details>

### Part 4 — Compile and link

The dedicated `uwp4me build` command is coming. Until then, this script does the same job. It also creates, on
demand, any link for a header spelling the SDK did not need itself.

<details>
<summary><b>build.sh</b> (click to expand)</summary>

```bash
cat > build.sh <<'EOF'
NAME="${1:-hello}"
S="$HOME/.uwp4me/sdk"
RES=$(clang -print-resource-dir)
CF="--target=x86_64-pc-windows-msvc -std=c++20 -fms-extensions -fms-compatibility -fms-compatibility-version=19.44 -fdelayed-template-parsing -fexceptions -fcxx-exceptions -D_MT -D_DLL -DUNICODE -D_UNICODE -DWINAPI_FAMILY=2 -D_WIN32_WINNT=0x0A00 -nostdinc -Wno-everything -isystem $RES/include -isystem $S/crt/include -isystem $S/sdk/include/ucrt -isystem $S/sdk/include/um -isystem $S/sdk/include/shared -isystem $S/sdk/include/winrt -isystem $S/sdk/include/cppwinrt"

: > build.log
ok=0; prev=""
for i in $(seq 1 80); do
  rm -f $NAME.obj
  out=$(clang++ $CF -c $NAME.cpp -o $NAME.obj 2>&1)
  if [ -s $NAME.obj ]; then ok=1; echo "compiled"; break; fi
  miss=$(echo "$out" | grep -oE "'[^']+' file not found" | head -n 1 | sed -E "s/'([^']+)'.*/\1/")
  if [ -z "$miss" ]; then echo "$out" | grep -E "error" | head -n 20; break; fi
  if [ "$miss" = "$prev" ]; then echo "could not resolve: $miss"; break; fi
  prev="$miss"
  b="${miss##*/}"; sub="${miss%/*}"; [ "$sub" = "$miss" ] && sub="."
  made=0
  for d in "$S/sdk/include/cppwinrt" "$S/sdk/include/winrt" "$S/sdk/include/um" "$S/sdk/include/shared" "$S/sdk/include/ucrt" "$S/crt/include"; do
    dir="$d/$sub"
    [ -d "$dir" ] || continue
    [ -e "$dir/$b" ] && continue
    f=$(find -L "$dir" -maxdepth 1 -iname "$b" | head -n 1)
    if [ -n "$f" ]; then ln -s "${f##*/}" "$dir/$b" && echo "link created: $miss" && made=1 && break; fi
  done
  if [ "$made" = 0 ]; then echo "not found in the SDK: $miss"; break; fi
done

if [ "$ok" = 1 ]; then
  echo "linking as a UWP app"
  lld-link $NAME.obj /out:$NAME.exe /subsystem:windows /machine:x64 /appcontainer /entry:wWinMainCRTStartup /opt:ref /dynamicbase /nxcompat "/libpath:$S/crt/lib/x64/store" "/libpath:$S/crt/lib/x64" "/libpath:$S/sdk/lib/ucrt/x64" "/libpath:$S/sdk/lib/um/x64" msvcrt.lib vcruntime.lib ucrt.lib msvcprt.lib WindowsApp.lib d3d11.lib dxgi.lib 2>&1 | head -n 25
  ls -l $NAME.exe
fi
EOF
sh build.sh hello
```

</details>

When it prints `hello.exe`, the app is compiled.

### Part 5 — Package and sign

<details>
<summary><b>pack.sh</b> (click to expand)</summary>

```bash
cat > genpng.go <<'EOF'
package main

import (
    "image"
    "image/color"
    "image/png"
    "os"
)

func gen(name string, size int) {
    img := image.NewRGBA(image.Rect(0, 0, size, size))
    for y := 0; y < size; y++ {
        for x := 0; x < size; x++ {
            img.Set(x, y, color.RGBA{30, 120, 200, 255})
        }
    }
    f, _ := os.Create(name)
    png.Encode(f, img)
    f.Close()
}

func main() {
    d := os.Args[1]
    gen(d+"/StoreLogo.png", 50)
    gen(d+"/Square44x44Logo.png", 44)
    gen(d+"/Square150x150Logo.png", 150)
}
EOF

cat > pack.sh <<'EOF'
NAME="${1:-hello}"
[ -s "$NAME.exe" ] || { echo "$NAME.exe not found, run: sh build.sh $NAME"; exit 1; }

mkdir -p Assets
[ -f Assets/StoreLogo.png ] || go run genpng.go Assets
rm -rf pkg && mkdir -p pkg/Assets
cp Assets/* pkg/Assets/
cp "$NAME.exe" "pkg/$NAME.exe"

cat > pkg/AppxManifest.xml <<EOM
<?xml version="1.0" encoding="utf-8"?>
<Package xmlns="http://schemas.microsoft.com/appx/manifest/foundation/windows10" xmlns:uap="http://schemas.microsoft.com/appx/manifest/uap/windows10" IgnorableNamespaces="uap">
  <Identity Name="UWP4Me.$NAME" Publisher="CN=UWP4Me Dev" Version="1.0.0.0" ProcessorArchitecture="x64"/>
  <Properties>
    <DisplayName>$NAME</DisplayName>
    <PublisherDisplayName>UWP4Me</PublisherDisplayName>
    <Logo>Assets\StoreLogo.png</Logo>
  </Properties>
  <Dependencies>
    <TargetDeviceFamily Name="Windows.Universal" MinVersion="10.0.17763.0" MaxVersionTested="10.0.26100.0"/>
    <PackageDependency Name="Microsoft.VCLibs.140.00" MinVersion="14.0.0.0" Publisher="CN=Microsoft Corporation, O=Microsoft Corporation, L=Redmond, S=Washington, C=US"/>
  </Dependencies>
  <Resources>
    <Resource Language="en-us"/>
  </Resources>
  <Applications>
    <Application Id="App" Executable="$NAME.exe" EntryPoint="$NAME.App">
      <uap:VisualElements DisplayName="$NAME" Description="Made with UWP4Me" BackgroundColor="#1E78C8" Square150x150Logo="Assets\Square150x150Logo.png" Square44x44Logo="Assets\Square44x44Logo.png"/>
    </Application>
  </Applications>
  <Capabilities>
    <Capability Name="internetClient"/>
  </Capabilities>
</Package>
EOM

if [ ! -f cert.pfx ]; then
  openssl req -x509 -newkey rsa:2048 -nodes -keyout key.pem -out cert.pem -days 365 -subj "/CN=UWP4Me Dev" -addext "basicConstraints=critical,CA:FALSE" -addext "keyUsage=critical,digitalSignature" -addext "extendedKeyUsage=critical,codeSigning"
  openssl pkcs12 -export -out cert.pfx -inkey key.pem -in cert.pem -passout pass:uwp4me
  openssl x509 -in cert.pem -outform der -out uwp4me.cer
fi

rm -f "$NAME.msix" "$NAME-signed.msix"
LD_LIBRARY_PATH="$HOME/.uwp4me/lib" "$HOME/.uwp4me/bin/makemsix" pack -d pkg -p "$NAME.msix"
osslsigncode sign -pkcs12 cert.pfx -pass uwp4me -h sha256 -in "$NAME.msix" -out "$NAME-signed.msix"
ls -l "$NAME-signed.msix" && cp "$NAME-signed.msix" /sdcard/Download/ && echo "copied to Download"
EOF
sh pack.sh hello
```

</details>

Two things worth knowing:

- The `Publisher` in the manifest (`CN=UWP4Me Dev`) **must match the certificate subject**, or the Xbox refuses the package.
- `cert.pfx` and `key.pem` are your **private signing key**. Never publish them. (The repo's `.gitignore` already ignores them.) The password in the script (`uwp4me`) is for development only.

### Part 6 — Install on the Xbox

1. On the phone's browser, open **`https://<XBOX-IP>:11443`** and accept the certificate warning (the console uses its own certificate).
2. Log in with the Device Portal username and password.
3. On **Home**, use **Add**, pick `hello-signed.msix` from the Download folder and install.
4. Open the app from **Dev Home** or from **My games & apps**.

If everything worked, a window opens and stays open (it is black: the hello app draws nothing yet).

> 💡 Dev Mode apps have memory limits. Per Microsoft's documentation, foreground **apps** get about 1 GB and **games** about 5 GB. Dev Home lets you choose the type when installing, so install games as **Game**. Check Microsoft's current docs, since these numbers can change.

### Next: a real game

The repository's roadmap includes ready-made templates. The **Snake** game used to validate UWP4Me draws with Direct3D 11
(`ClearView`, no shaders needed), reads the Xbox gamepad with `Windows.Gaming.Input` and runs a simple game loop.
It is built with exactly the same `build.sh` and `pack.sh`, just with `-ld3d11 -ldxgi` libs already in the script.

---

## 🛟 Troubleshooting

| Symptom | What to do |
| --- | --- |
| `uwp4me doctor` shows red items | Each failure prints a `fix:` hint. Most are solved by `uwp4me setup` or `uwp4me sdk`. |
| Termux gets killed in the middle of a step | Plug in the charger, run `termux-wake-lock` and keep the screen on. Heavy steps can be repeated: downloads resume and finished work is cached. |
| `unexpected EOF` or `HTTP 416` while downloading | Just run `uwp4me sdk` again. Downloads resume and every file is checked with SHA-256. |
| `'something.h' file not found` | `build.sh` creates the missing link by itself. If it stops with `not found in the SDK`, please open an issue with the name. |
| `makemsix` build fails | Look at `~/.uwp4me/msix-build.log` and open an issue with its last 30 lines. |
| Pasting a long script in Termux breaks it | Termux can cut very long pastes. Paste in smaller pieces and check the file with `tail`. |
| The Xbox refuses to install the package | Check that the manifest `Publisher` equals the certificate subject, and that Device Portal is on. Copy the exact error text into an issue. |
| The app opens and closes immediately | Open an issue with the app type (App or Game) and what you did. The `EntryPoint` and framework dependency in the manifest are the first suspects. |
| Want to start over | `uwp4me clean --all`, then `uwp4me` again. |

---

## 🗺️ Roadmap

- [x] Build `makemsix` on Android with a pinned, patched MSIX Packaging SDK
- [x] Windows SDK downloader with license prompt, SHA-256 checks and resume
- [x] Automatic file-name case fixing
- [x] `doctor` and `clean`
- [ ] `init` with templates (hello world, snake)
- [ ] `build`, `pack` and `run` as single commands
- [ ] `deploy` straight to the Device Portal from the terminal
- [ ] Slim SDK: keep only what UWP needs
- [ ] Pinned SDK versions for fully reproducible setups
- [ ] Prebuilt binaries in GitHub Releases
- [ ] On-screen text and sound helpers for games

---

## ⚖️ Legal and licensing

- **UWP4Me** itself is released under the **[MIT License](LICENSE)**.
- UWP4Me **does not contain or redistribute any Microsoft file.** The Windows SDK and the Visual C++ runtime are downloaded **directly from Microsoft's servers** on your device, and only after **you accept Microsoft's license terms** (the tool asks, and remembers your answer).
- The [MSIX Packaging SDK](https://github.com/microsoft/msix-packaging) (MIT, by Microsoft) is fetched from its official repository at a pinned commit and patched locally with the small patch included in `patches/`.
- *UWP4Me is an independent community project. It is **not affiliated with, endorsed by or sponsored by Microsoft.** Xbox, Windows and Visual Studio are trademarks of Microsoft.*
- Use Developer Mode and this tool only on consoles you own, and follow Microsoft's Developer Mode terms.

---

## 🙏 Credits

- [Microsoft MSIX Packaging SDK](https://github.com/microsoft/msix-packaging): the packager at the heart of the project.
- [LLVM](https://llvm.org): `clang` and `lld`.
- [osslsigncode](https://github.com/mtrojnar/osslsigncode): signing `.msix` packages.
- [msitools](https://wiki.gnome.org/msitools): reading and extracting the SDK installers.
- [xwin](https://github.com/Jake-Shadle/xwin) by Jake Shadle: the inspiration for downloading the Windows SDK without Windows.
- [Termux](https://termux.dev): what makes a phone a real dev machine.

---

## 🤝 Contributing

Issues and pull requests are welcome, especially:

- reports from **other phones and architectures** (64-bit ARM, x86 emulators),
- tests on **other Xbox models** (One, Series X),
- new **templates** and examples,
- help **slimming the SDK**.

When reporting a problem, please paste the output of `uwp4me doctor`.

---

<details>
<summary>🇧🇷 <b>Resumo em português</b></summary>

**UWP4Me** é uma ferramenta de terminal que **compila, empacota e assina apps de Xbox direto do celular Android (Termux)**, sem PC e sem Visual Studio.

- O compilador é o `clang` fazendo cross-compile pra Windows x64, com o Windows SDK baixado **dos servidores da própria Microsoft** (depois de você aceitar a licença).
- O empacotador é o MSIX Packaging SDK da Microsoft, compilado no seu celular com um pequeno patch.
- A assinatura usa o `osslsigncode` e a instalação é pelo **Device Portal** do Xbox em Dev Mode.

**Começar:**

```bash
pkg install -y git golang
git clone https://github.com/muriloprimata/uwp4me.git
cd uwp4me && go build -o uwp4me ./cmd/uwp4me
./uwp4me          # primeira execução: instala tudo (10 a 15 min)
./uwp4me sdk      # aceita a licença e baixa o SDK
./uwp4me doctor   # confere se está tudo certo
```

Depois siga o **Tutorial** acima (Partes 1 a 6) pra fazer o seu primeiro app rodar no Xbox. Já foi testado com um hello world e um jogo da cobrinha completo num **Xbox Series S**.

</details>

---

<div align="center">

Made with ☕ and a phone by **Bieseche** ([@muriloprimata](https://github.com/muriloprimata))

⭐ If UWP4Me helped you, a star on the repo means a lot.

</div>
