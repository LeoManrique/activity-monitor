# IMPLEMENTATION_MEMORY.md — Activity Monitor (Memory Usage MVP)

> Wails 3 desktop app showing memory usage per program, grouped by app name.
> Focus: Go backend logic. Frontend: minimal/functional.
> Each phase produces a runnable app you can `wails3 dev` and see results.

---

## Phase 1 — Setup & Hello World

### 1.1 Prerequisites & Environment Setup

- Install Go 1.24+ from https://go.dev/dl/ if you do not already have it. Verify with `go version`.
- Install the Wails 3 CLI by running: `go install github.com/wailsapp/wails/v3/cmd/wails3@latest`
- Make sure your Go bin directory is on your PATH. The default is `~/go/bin`. If `wails3` is not found after installing, add `export PATH="$HOME/go/bin:$PATH"` to your shell profile (~/.zshrc on macOS) and restart your terminal.
- On macOS you need Xcode Command Line Tools. Run `xcode-select --install` if you have not already.
- Verify everything is in order by running: `wails3 doctor`. This checks for Go, platform toolchains, and any missing dependencies. Fix anything it reports before continuing.

> **Go concept — `go install`:** This command downloads a Go module from the internet, compiles it, and places the resulting binary in your GOBIN directory. It is the standard way to install CLI tools written in Go. Unlike npm's global installs, there is no central registry — the argument is a module path that points directly to a Git repository.

### 1.2 Scaffold the Project

- Run `wails3 init -n activity-monitor -t svelte` from the parent directory where you want the project to live. The `-n` flag sets the project name and directory. The `-t svelte` flag selects the Svelte + Vite frontend template. (Run `wails3 init -l` to see all available templates if you are curious.)
- After the command finishes, `cd activity-monitor` and look at what was generated. Here is every file and folder and what it does:

| File / Folder | Purpose |
|---|---|
| `go.mod` | Declares this project as a Go module and lists its dependencies (including Wails v3). The Go equivalent of package.json. |
| `go.sum` | Lock file with checksums for every dependency. Do not edit by hand. |
| `main.go` | Entry point for the Go application. Creates the Wails app, registers services, configures the window, and calls `app.Run()`. |
| `greetservice.go` | A sample service struct (`GreetService`) with one exported method (`Greet`). This is the example you will replace with your own service in Phase 2. |
| `frontend/` | The entire Svelte frontend project (Vite-based). Contains HTML, CSS, JS, and Svelte components. |
| `frontend/index.html` | The HTML shell that Wails loads into the native webview. |
| `frontend/src/` | Svelte source files — `main.js` (entry point), `App.svelte` (root component), and styles. |
| `frontend/src/assets/` | Static assets like images and fonts used by the frontend. |
| `frontend/package.json` | Node dependencies for the frontend (Svelte, Vite, etc.). |
| `frontend/vite.config.js` | Vite bundler configuration. |
| `frontend/bindings/` | Auto-generated JavaScript (or TypeScript) binding files. These are created by `wails3 generate bindings` and give your frontend typed functions that call your Go methods. You should not edit these by hand. |
| `Taskfile.yml` | Task runner configuration (similar to a Makefile). Defines the `dev`, `build`, and `generate` commands that `wails3` uses internally. |

- **Key insight: the `frontend/` folder gets embedded into the final binary.** In your `main.go`, you will see a Go `embed` directive: `//go:embed frontend/dist`. At build time, Go's embed package takes the built frontend output and bakes it directly into the compiled binary. Unlike Go source code which is compiled to machine code, these frontend assets (HTML, CSS, JS) are bundled as-is — they are served from memory at runtime. During development with `wails3 dev`, the frontend is served by Vite's dev server instead, which gives you hot-reload: edit a Svelte file, save, and see the change instantly without recompiling Go.
- Run `cd frontend && pnpm install && cd ..` to install the frontend's Node dependencies. Some templates do this automatically, but verify by checking that `frontend/node_modules/` exists.

### 1.3 The Wails 3 Mental Model

The core idea of Wails 3 is: **your Go struct IS your API**. Here is how data flows from Go to the screen:

1. **You write a Go struct** with exported methods. For example, a struct called `MemoryService` with a method called `GetMemoryUsage` that returns a slice of data. This is just normal Go — no annotations, no decorators, no special interfaces required.
2. **You register the struct** with the Wails application by adding `application.NewService(&YourStruct{})` to the `Services` slice in `application.Options` inside `main.go`.
3. **Wails generates JavaScript bindings** (via `wails3 generate bindings`). For each exported method on your registered service, Wails creates a matching JavaScript function in `frontend/bindings/`. That function handles the IPC call to Go behind the scenes.
4. **Your Svelte frontend imports and calls those generated functions.** The call crosses from the webview into Go, executes your method, and the return value comes back as a JavaScript object (JSON-serialized automatically). If your Go method returns an error as the second value, the JS function returns a rejected Promise.

The "contract" between your frontend and backend is your struct and its methods. Change a method signature in Go, re-run bindings generation, and the frontend sees the new shape immediately. There is no REST API, no HTTP server, no manual JSON marshaling — Wails handles all of that.

> **Go concept — exported vs unexported:** In Go, a name that starts with an uppercase letter (like `GetMemoryUsage`) is "exported" — visible outside the package, and thus visible to Wails for binding. A name starting with lowercase (like `parseProcessLine`) is "unexported" — private to the package. There is no `public`/`private` keyword; the casing of the first letter is the access control mechanism.

### 1.4 First Run & Devtools

- From the project root (`activity-monitor/`), run: `wails3 dev`
- This does three things simultaneously: it starts the Vite dev server for the frontend, compiles and runs the Go backend, and opens a native macOS window with a webview pointing at the Vite dev server.
- You should see the default Wails welcome page in a native window. It will have a text input and a "Greet" button — this is the sample `GreetService` in action.
- To open browser developer tools (console, network, elements inspector) inside the Wails window, right-click anywhere in the window and select **Inspect Element** from the context menu, or press **Cmd+Option+I**. The devtools panel will appear inside or alongside the window, exactly like Chrome DevTools. You will use this constantly to debug JavaScript, inspect network calls to your Go backend, and view console output.
- Try it now: open devtools, go to the Console tab, and type `console.log("hello from devtools")`. You should see the output. This confirms you have a working debugging environment.
- Leave `wails3 dev` running while you work. When you edit Go files, it will recompile automatically. When you edit Svelte/JS/CSS files, changes appear instantly via Vite hot-reload.
- To stop the dev server, press `Ctrl+C` in the terminal where `wails3 dev` is running.

**Checkpoint:**

- You see the default Wails welcome page in a native macOS window with the "Greet" button.
- You can type a name, click Greet, and see a greeting message — confirming the Go-to-JS binding pipeline works end to end.
- You can open browser devtools inside the Wails window with right-click > Inspect Element or Cmd+Option+I.
- Running `console.log("test")` in the devtools Console tab prints output, confirming your debugging environment is ready.

---

## Phase 2 — Your First Service (Hardcoded Data)

### 2.1 Create the `AppMemoryGroup` Struct

- Create a new file called `memory.go` in the project root (same directory as `main.go`). All of the memory-related types and logic will live in this file.
- At the top, declare `package main` — this must match the package name used in `main.go`.
- Define a struct named `AppMemoryGroup` with three fields: `Name` of type `string`, `MemoryKB` of type `int64`, and `ProcessCount` of type `int`.
- Add a JSON struct tag to each field so the frontend receives clean lowercase keys. The tag for `Name` should be `` `json:"name"` ``, for `MemoryKB` it should be `` `json:"memoryKB"` ``, and for `ProcessCount` it should be `` `json:"processCount"` ``.
- This struct is the data contract between your Go backend and your Svelte frontend. Every field here will appear as a property on the JavaScript objects your frontend receives.

> **Go concept — struct tags:** In Go, you can attach metadata to struct fields using backtick-delimited strings after the type. The `json:"name"` tag tells Go's JSON encoder (and Wails' binding generator) to use `name` as the key instead of the Go field name `Name`. Without tags, JSON keys would match the Go field names exactly (uppercase), which is unconventional in JavaScript. Tags are not executable code — they are metadata read at runtime via reflection.

### 2.2 Create `MemoryService` Struct

- In the same `memory.go` file, below your `AppMemoryGroup` struct, define a new struct called `MemoryService` with no fields — just an empty struct body.
- This struct is the "service" that Wails will bind to the frontend. You will attach methods to it, and Wails will make each exported method callable from JavaScript.
- You do not need any imports for this struct definition alone.

> **Go concept — methods on structs:** In Go, you cannot just write standalone functions and register them with Wails. Instead, you define methods that have a "receiver" — a struct they are attached to. This is Go's version of a class with methods. The struct does not need any fields to have methods; an empty struct works fine as a method namespace. Later, when you have state to manage (like a database connection), you would add fields to the struct.

### 2.3 Add `GetMemoryUsage()` Method — Hardcoded

- Define a method named `GetMemoryUsage` on `MemoryService` (using a pointer receiver: the receiver type should be `*MemoryService`). It takes no arguments and returns a `[]AppMemoryGroup` — a slice of your struct.
- Inside the method, build a slice literal containing four hardcoded `AppMemoryGroup` values and return it. Use these fake entries:
  - "Google Chrome" with MemoryKB 1_500_000 and ProcessCount 23
  - "Slack" with MemoryKB 800_000 and ProcessCount 5
  - "Firefox" with MemoryKB 600_000 and ProcessCount 12
  - "Finder" with MemoryKB 150_000 and ProcessCount 2
- Use the underscore numeric separator (`1_500_000`) to make large numbers readable — Go allows this in integer literals.
- Make sure you import nothing extra for this step. The method is pure data, no external packages needed.

> **Go concept — slice literals:** A slice literal looks like `[]TypeName{ item1, item2, item3 }`. Each item is a struct literal like `AppMemoryGroup{Name: "Chrome", MemoryKB: 1500000, ProcessCount: 23}`. The trailing comma after the last item is mandatory in Go if the closing brace is on a new line. This is different from JavaScript/Python where trailing commas are optional.

### 2.4 Register the Service in `main.go`

- Open `main.go`. Find the `application.Options` struct where `Services` is defined. You should see the existing `GreetService` already registered there.
- Add a second entry to the `Services` slice: `application.NewService(&MemoryService{})`. Place it right after the existing `GreetService` entry. The `&` is important — it creates a pointer to your struct, which Wails requires so it can call your pointer-receiver methods.
- You do not need to add any new imports — `MemoryService` is in the same `main` package and `application` is already imported.
- You can leave the `GreetService` registration in place for now. It does no harm and you can use it to verify the app still starts correctly. You will remove it later when you clean up.

> **Go concept — pointer with `&`:** The `&` operator takes the address of a value, giving you a pointer. `&MemoryService{}` creates a new `MemoryService` value and immediately takes its address. Wails needs a pointer because your method has a pointer receiver (`*MemoryService`). If you passed `MemoryService{}` without `&`, Go would not find the method. This pointer/value distinction does not exist in languages like Python or JavaScript, where objects are always references.

### 2.5 Generate Bindings

- Stop `wails3 dev` if it is running (Ctrl+C).
- From the project root, run: `wails3 generate bindings`
- You should see output indicating it processed your services and methods. It will report finding `MemoryService` and its `GetMemoryUsage` method.
- The generated files are written to `frontend/bindings/`. Look inside that directory — you will find a subfolder structure based on your module name. Navigate to `frontend/bindings/` and look for a folder named `activity-monitor` (derived from your Go module path), then inside it a file related to `memoryservice`.
- Open the generated JavaScript file for `MemoryService`. You will see a function named `GetMemoryUsage` that internally calls `window.wails.Call(...)`. This is the function your Svelte code will import and call. You do not need to edit this file — it is regenerated every time you run the bindings command.
- Note: `wails3 dev` also regenerates bindings automatically when it detects Go file changes. But running the command manually is useful when you want to inspect the output or debug import paths.

### 2.6 Build a Minimal Frontend Table

- Open `frontend/src/App.svelte`. This is the root Svelte component that currently shows the default Wails greeting UI. You will replace its contents entirely.
- Delete everything in `App.svelte` and replace it with the following Svelte component. This component calls your Go method on mount, stores the result, and renders it as a table:

```svelte
<script>
  import { onMount } from "svelte";
  import { GetMemoryUsage } from "../bindings/github.com/LeoManrique/activity-monitor/memoryservice";

  let groups = [];

  onMount(async () => {
    try {
      groups = await GetMemoryUsage();
    } catch (err) {
      console.error("Failed to fetch memory usage:", err);
    }
  });
</script>

<main>
  <h1>Memory Usage</h1>
  <table>
    <thead>
      <tr>
        <th>Application</th>
        <th>Memory (KB)</th>
        <th>Processes</th>
      </tr>
    </thead>
    <tbody>
      {#each groups as group}
        <tr>
          <td>{group.name}</td>
          <td>{group.memoryKB}</td>
          <td>{group.processCount}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</main>
```

- The import path `../bindings/github.com/LeoManrique/activity-monitor/memoryservice` matches the directory structure that `wails3 generate bindings` created. If your import path does not match, check the actual folder names inside `frontend/bindings/` and adjust accordingly.
- The `onMount` function is Svelte's lifecycle hook — it runs once when the component first renders in the DOM. Since `GetMemoryUsage` returns a Promise (all Wails binding calls are async), you use `await` inside an `async` callback.
- The property names in the template (`group.name`, `group.memoryKB`, `group.processCount`) match the JSON tags you defined on your Go struct. If you used different JSON tags, adjust these to match.
- Remember: frontend files are embedded into the final binary at build time, but during `wails3 dev` they are served by Vite with hot-reload. You can edit this Svelte file and see changes instantly without restarting the Go backend.

### 2.7 Run & Verify

- From the project root, run: `wails3 dev`
- The native window should open and display a table with four rows: Google Chrome, Slack, Firefox, and Finder, showing their hardcoded memory and process count values.
- If you see a blank window or errors, open devtools (Cmd+Option+I) and check the Console tab. Common issues at this stage:
  - **"GetMemoryUsage is not a function"** or **import errors**: The binding import path is wrong. Check the actual folder names inside `frontend/bindings/` and correct the import in `App.svelte`.
  - **"wails.Call is not defined"**: The Wails runtime is not loaded. Make sure you are running via `wails3 dev` and not just `npm run dev`.
  - **Empty table (no rows)**: The call succeeded but returned empty. Check that `GetMemoryUsage` actually returns the hardcoded slice and that the method name starts with an uppercase letter (exported).
- If you make changes to `App.svelte`, they will appear immediately without restarting. If you change Go code (like adding a new fake entry), the dev server will automatically recompile and restart the backend.

**Checkpoint:** Hardcoded Go data showing in the browser. The full Go-to-bindings-to-JS-to-DOM pipeline is proven.

- You see a table with four rows: Google Chrome (1500000 KB, 23 processes), Slack (800000 KB, 5 processes), Firefox (600000 KB, 12 processes), and Finder (150000 KB, 2 processes).
- Open devtools and type `GetMemoryUsage` in the Console — if the bindings are loaded correctly, it should be reachable from the module scope (though direct console access depends on bundling).
- Try changing one of the hardcoded values in `memory.go` (for example, change Chrome's MemoryKB to 2000000). After `wails3 dev` recompiles, refresh the window and confirm the new value appears in the table. This proves the full round-trip: Go change, recompile, binding call, and frontend render.

---

## Phase 3 — Running a System Command from Go

### 3.1 Understanding `os/exec`

- Go's standard library includes the `os/exec` package, which lets you run external commands (like terminal commands) from within your Go program. This is the Go equivalent of Python's `subprocess.run()` or Node's `child_process.execSync()`.
- The two key pieces you will use are `exec.Command()` and `.Output()`. The function `exec.Command(name, args...)` creates a command object — the first argument is the program name, and any remaining arguments are passed as separate strings (not one big string like you would type in a shell). For example, to represent the terminal command `ls -la /tmp`, you would call `exec.Command("ls", "-la", "/tmp")` — three separate arguments, not one string containing spaces.
- Calling `.Output()` on the command object runs the subprocess, waits for it to finish, and returns everything it printed to stdout as a `[]byte` (a byte slice). It also returns an `error` as a second value, which will be non-nil if the command failed to execute or exited with a non-zero status code.
- You convert the `[]byte` result to a `string` using Go's built-in `string()` conversion function. This is a zero-cost type conversion, not a function call that copies data — Go strings and byte slices share the same underlying representation (UTF-8 bytes).

> **Go concept — multiple return values:** Go functions can return more than one value. The convention for operations that can fail is to return `(result, error)`. The caller must handle both: check if `error` is not nil before using the result. This replaces exceptions — Go does not have try/catch. Every error is an explicit value you must check. If you ignore the error return value, your program may crash on nil data later.

### 3.2 The `ps` Command We'll Use

- Before writing any Go code, run this exact command in your terminal right now: `ps -axo pid,rss,comm`
- Study the output. You will see a header line followed by hundreds of rows. Each row has three columns: a process ID (PID), a memory value (RSS in kilobytes), and a command path. The output looks roughly like this (your values will differ):

```
  PID   RSS COMM
    1  23456 /sbin/launchd
   87  12800 /usr/libexec/logd
  441 854320 /Applications/Google Chrome.app/Contents/MacOS/Google Chrome
```

- Here is what each flag means:
  - `-a` — show processes belonging to all users, not just you.
  - `-x` — include processes that are not attached to a terminal (background daemons, services, menu bar apps). Without this flag, you would miss most of the system.
  - `-o pid,rss,comm` — output only these three columns in this exact order, instead of the default wide format. `pid` is the process ID, `rss` is Resident Set Size in kilobytes (the actual physical memory the process is using right now), and `comm` is the command name (the executable path).
- Notice that the `comm` column often shows a full path like `/Applications/Slack.app/Contents/MacOS/Slack`. In Phase 4 you will extract just the last component (`Slack`), but for now you only need the raw output.
- Also notice that some processes have very small RSS values (a few hundred KB) while browsers and Electron apps can show hundreds of thousands. The numbers are in kilobytes — so `854320` means roughly 834 MB.
- Keep this terminal output visible. You will compare it against your Go program's output in a moment to verify correctness.

### 3.3 Write `getProcessList() (string, error)`

- Create a new file called `processes.go` in the project root (same directory as `main.go` and `memory.go`). This file will hold all the process-fetching and parsing logic. Keeping it separate from `memory.go` (which holds your structs and service) is a matter of organization — they are all in the same `main` package.
- At the top of the file, declare `package main`.
- Add an import block that includes `"os/exec"`.
- Define an unexported function named `getProcessList`. It takes no arguments and returns two values: a `string` and an `error`. It is unexported (lowercase first letter) because it is an internal helper — only your Go code will call it, not the frontend.
- Inside the function, use `exec.Command` to create a command for `"ps"` with three arguments: `"-axo"`, `"pid,rss,comm"`. Note that `-axo` and `pid,rss,comm` are two separate arguments — do not combine them into one string.
- Call `.Output()` on the command to execute it and capture stdout. This returns a `[]byte` and an `error`.
- If the error is not nil, return an empty string and the error. This handles cases where `ps` is not found or fails to run.
- If there is no error, convert the `[]byte` to a `string` and return it along with a `nil` error.
- That is the entire function — it should be roughly six to eight lines of actual code (not counting the package declaration and imports).

> **Go concept — `[]byte` vs `string`:** In Go, `string` and `[]byte` are closely related but distinct types. A `string` is immutable (you cannot change individual characters), while `[]byte` is mutable. The `os/exec` package returns `[]byte` because it is reading raw bytes from a subprocess. You convert to `string` with `string(myBytes)` when you want to do text operations like splitting by newlines. This conversion is very cheap — Go does not re-encode or validate the data.

### 3.4 Test It in Isolation

- Open `main.go`. Find the `main` function — it currently creates the Wails application and calls `app.Run()`.
- Temporarily add test code at the very beginning of the `main` function, before any Wails setup code. Call `getProcessList()` and store the two return values (the output string and the error).
- Check if the error is not nil. If it is, print the error using `fmt.Println` and return early from main.
- If there is no error, print the first 500 characters of the output string using `fmt.Println`. To get the first 500 characters, slice the string with bracket notation — a string in Go can be sliced like `s[:500]` to get the first 500 bytes. This avoids flooding your terminal with hundreds of lines.
- You will need to add `"fmt"` to the imports in `main.go` if it is not already there.
- Run the app with `go run .` from the project root. You do not need `wails3 dev` for this test — `go run .` compiles and runs the Go code directly. The Wails window may still open, but focus on the terminal output.
- You should see the first chunk of `ps` output printed to your terminal — a header line (`PID   RSS COMM`) followed by several process rows. Compare this against the output you saw when you ran `ps -axo pid,rss,comm` directly in your terminal in section 3.2. The content should be essentially the same (process lists change moment to moment, so exact values will differ, but the format and structure should match).
- If you see an error message instead, double-check that you spelled the `exec.Command` arguments correctly — the most common mistake is combining the flags into a single string instead of separate arguments.
- **Once you have confirmed the output looks correct, delete all the test code you just added to `main.go`.** Remove the `getProcessList()` call, the `fmt.Println` lines, and the early return. Also remove the `"fmt"` import if nothing else in `main.go` uses it (Go will refuse to compile if you have an unused import). Leave `main.go` exactly as it was before this test. The `getProcessList` function in `processes.go` stays — you will use it in the next phase.

> **Go concept — unused imports are errors:** Unlike most languages where unused imports are just warnings, Go treats them as compile errors. If you import `"fmt"` but do not use any function from it, your program will not compile. This strictness keeps Go codebases clean but can be annoying during development. A common workaround is to add `_ = fmt.Println` as a throwaway usage, but the better habit is to simply add and remove imports as needed.

**Checkpoint:** You can run a system command from Go and get its output as a string. This is a reusable skill.

- Running `go run .` with the temporary test code prints process data to your terminal that matches the format of `ps -axo pid,rss,comm`: a header line followed by rows of PID, RSS, and command path columns.
- The output contains real process names you recognize from your running system — you should see entries for things like `launchd`, `WindowServer`, or whatever applications you have open.
- After removing the test code from `main.go`, running `wails3 dev` still works and shows the same hardcoded table from Phase 2 — confirming you did not accidentally break anything. The `getProcessList` function exists in `processes.go` and compiles without errors, but nothing calls it yet.

---

## Phase 4 — Parsing the Output into Structs

### 4.1 Define `ProcessInfo` Struct

- Open `processes.go` — the file you created in Phase 3 where `getProcessList` already lives. You will add all parsing types and functions to this same file.
- Define a struct named `ProcessInfo` with three fields: `PID` of type `int`, `RSS` of type `int64`, and `Command` of type `string`.
- This struct does not need JSON tags because it will never be sent directly to the frontend. It is an intermediate representation — raw data from `ps` that you will later group into `AppMemoryGroup` structs (which do have JSON tags). Think of `ProcessInfo` as an internal data transfer object that only your Go code touches.
- `RSS` is `int64` rather than `int` because memory values can be large on systems with lots of RAM. While `int` would technically work on 64-bit macOS (where `int` is 64 bits), using `int64` makes the intent explicit and portable.

> **Go concept — unexported structs:** Even though `ProcessInfo` starts with an uppercase letter (making it exported), it will only be used within the `main` package in this project. Uppercase naming here is a stylistic choice — it makes the struct name readable and consistent with `AppMemoryGroup`. Since this is a `main` package (an executable, not a library), no external code can import it regardless of casing. In a library package, you would use lowercase (`processInfo`) to keep it private.

### 4.2 Split Output into Lines

- Look at the raw `ps` output you examined in Phase 3. The first line is always a header: `  PID   RSS COMM`. Every line after that is a process entry. Some lines at the end may be empty (a trailing newline produces an empty string after splitting).
- The `strings` package in Go's standard library provides `strings.Split(s, sep)`, which splits a string into a slice of substrings at every occurrence of the separator. Splitting on `"\n"` (a newline character) gives you one string per line of output.
- After splitting, the element at index 0 is the header line. You must skip it — it contains column labels, not data. You must also skip any empty strings in the slice, which result from trailing newlines or blank lines in the output.
- You will use this splitting logic inside the `parseAllProcesses` function (section 4.5), not as a standalone function. Understanding the structure of the output first will make writing the parser straightforward.

### 4.3 Parse a Single Line

- Add the following imports to the import block at the top of `processes.go`: `"strconv"`, `"strings"`, and `"path/filepath"`. You already have `"os/exec"` from Phase 3. Combine them all in a single parenthesized import block.
- Define an unexported function named `parseProcessLine`. It takes a single `string` parameter (one line of `ps` output) and returns two values: a `ProcessInfo` and an `error`.
- Inside the function, use `strings.Fields()` to split the line into a slice of strings. This function is different from `strings.Split` — it splits on any amount of whitespace (spaces, tabs) and automatically ignores leading/trailing whitespace. This is exactly what you need because `ps` output uses variable-width spacing to align columns.
- `strings.Fields` on a line like `"  441 854320 /Applications/Google Chrome.app/Contents/MacOS/Google Chrome"` produces `["441", "854320", "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"]` — wait, that is not right. It actually produces `["441", "854320", "/Applications/Google", "Chrome.app/Contents/MacOS/Google", "Chrome"]` because Fields splits on every whitespace boundary. This matters: the command path can contain spaces. You need at least 3 fields (PID, RSS, and at least one part of the command), and everything from field index 2 onward is the command.
- Check that the fields slice has at least 3 elements. If it has fewer, return an empty `ProcessInfo` and an error (use `fmt.Errorf` with a descriptive message like "expected at least 3 fields, got %d"). Add `"fmt"` to your imports.
- Parse the first field (index 0) as the PID using `strconv.Atoi`, which converts a string to an `int`. If it returns an error, return an empty `ProcessInfo` and that error.
- Parse the second field (index 1) as the RSS using `strconv.ParseInt`. This function takes three arguments: the string to parse, the base (use 10 for decimal), and the bit size (use 64). It returns an `int64` and an `error`. If it returns an error, return an empty `ProcessInfo` and that error.
- Rejoin all fields from index 2 onward into a single string using `strings.Join(fields[2:], " ")`. This reconstructs the full command path even if it contained spaces. The `fields[2:]` syntax creates a sub-slice starting at index 2 through the end — this is Go's slice expression, similar to Python's list slicing.
- Store PID, RSS, and the rejoined command string in a `ProcessInfo` struct and return it with a `nil` error. You will handle the command name cleanup in the next section before storing it.

> **Go concept — `strconv.Atoi` vs `strconv.ParseInt`:** Go provides two families of string-to-number conversions. `Atoi` (ASCII to integer) is a shorthand that always parses as base-10 into an `int`. `ParseInt` gives you control over the base and the bit size of the result. You use `Atoi` for PID (which fits in a regular `int`) and `ParseInt` for RSS (which you want as `int64`). Both return an error if the string is not a valid number.

### 4.4 Handle the Command Name

- After you rejoin the command fields in `parseProcessLine` (the step above), the result is often a full filesystem path like `/Applications/Slack.app/Contents/MacOS/Slack` or `/usr/libexec/logd`. You want just the last component — `Slack` or `logd` — because that is the human-readable application name you will group by in Phase 5.
- Use `filepath.Base()` from the `"path/filepath"` package (already in your imports from section 4.3). This function takes a path string and returns the last element — the filename. For example, `filepath.Base("/usr/bin/python3")` returns `"python3"`, and `filepath.Base("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome")` returns `"Google Chrome"`.
- Apply `filepath.Base()` to the rejoined command string before storing it in the `Command` field of your `ProcessInfo` struct. This means the `Command` field will contain the short application name, not the full path.
- Note: `filepath.Base` handles edge cases for you — if the string is empty it returns `"."`, and if it is just `"/"` it returns `"/"`. These edge cases will be handled later in Phase 9 when you deal with unusual process names.

### 4.5 Write `parseAllProcesses(output string) []ProcessInfo`

- Define an unexported function named `parseAllProcesses`. It takes one `string` parameter (the entire raw output from `ps`) and returns a `[]ProcessInfo` — a slice of `ProcessInfo` structs. This function does not return an error; it handles bad lines internally by skipping them.
- Start by splitting the output string into lines using `strings.Split(output, "\n")`.
- Create an empty slice of `ProcessInfo` to collect results. You can declare it as `var result []ProcessInfo` — in Go, this creates a nil slice, which behaves identically to an empty slice for appending purposes.
- Iterate over the lines using a `for` loop with `range`. The `range` keyword over a slice gives you two values on each iteration: the index and the value. You need the index to skip the header line at index 0.
- For each line, skip it if the index is 0 (the header). Also skip the line if, after trimming whitespace with `strings.TrimSpace`, it is an empty string.
- For each remaining line, call `parseProcessLine` and receive the `ProcessInfo` and `error` return values. If the error is not nil, skip that line — do not add it to results, and do not crash. Some `ps` output lines may have unexpected formats (zombie processes, kernel threads), and silently skipping them is the correct behavior for a monitoring tool.
- If `parseProcessLine` succeeded (error is nil), append the `ProcessInfo` to your result slice using Go's built-in `append` function: `result = append(result, info)`. Note that `append` returns a new slice — you must assign it back. This is because slices in Go are backed by arrays, and appending may allocate a new, larger array.
- After the loop, return the result slice.

> **Go concept — `append` must be reassigned:** Unlike list.append() in Python or Array.push() in JavaScript which mutate the original, Go's `append` returns a new slice value that you must assign back to the variable. If you write `append(result, info)` without `result = `, the appended element is lost. This is because a Go slice is a small struct (pointer, length, capacity) passed by value. The `append` function may return a slice pointing to a completely new underlying array if the old one ran out of capacity.

### 4.6 Test the Parser

- Open `main.go`. Just as you did in Phase 3, add temporary test code at the very beginning of the `main` function, before any Wails setup code.
- Call `getProcessList()` to get the raw output string. Check the error — if it is not nil, print it with `fmt.Println` and return.
- Pass the output string to `parseAllProcesses()` to get a `[]ProcessInfo`.
- Print the total number of parsed processes using `fmt.Println` — something like "Parsed X processes" where X is the length of the slice. Use the built-in `len()` function to get the slice length.
- Print the first 5 entries of the result slice using a `for` loop over `result[:5]`. For each `ProcessInfo`, print all three fields (PID, RSS, Command) using `fmt.Printf` with a format string like `"PID=%d  RSS=%d KB  CMD=%s\n"`. The `%d` verb formats integers and `%s` formats strings.
- Run `go run .` from the project root. You should see output like:

```
Parsed 487 processes
PID=1  RSS=23456 KB  CMD=launchd
PID=87  RSS=12800 KB  CMD=logd
PID=141  RSS=4200 KB  CMD=UserEventAgent
PID=143  RSS=8960 KB  CMD=fseventsd
PID=183  RSS=6780 KB  CMD=systemstats
```

- Your exact numbers and process names will differ, but verify three things: (1) PIDs are reasonable positive integers, (2) RSS values are positive numbers in the kilobyte range, and (3) command names are short names like `launchd` or `Google Chrome` — not full paths like `/sbin/launchd`. If you see full paths, `filepath.Base()` is not being applied correctly.
- Compare your output against running `ps -axo pid,rss,comm` directly in another terminal window. Pick a few PIDs from the Go output and find them in the terminal output. The PID and RSS should match (RSS may differ slightly due to timing), and the command should be the basename of the path shown by `ps`.
- **Once you have verified the output, delete all the temporary test code from `main.go`** — the `getProcessList` call, the `parseAllProcesses` call, the print statements, and the early return. Remove the `"fmt"` import from `main.go` if nothing else uses it. Leave `main.go` exactly as it was. All the new code in `processes.go` stays.

**Checkpoint:** Raw `ps` output → structured `[]ProcessInfo`. You've practiced string parsing, type conversion, and error handling — all pure Go stdlib.

---

## Phase 5 — Grouping Processes by Application

### 5.1 The Grouping Strategy

- Right now you have a flat list of `ProcessInfo` structs — one per process. But many processes share the same application name: Google Chrome alone can have 20+ entries (one per tab, one for the GPU process, one for the main browser, etc.). The goal of this phase is to collapse all processes that share the same `Command` name into a single `AppMemoryGroup`, summing their memory and counting how many processes contributed.
- The data structure you will use is a Go map with type `map[string]*AppMemoryGroup`. The key is the process name (the `Command` field from `ProcessInfo`), and the value is a pointer to an `AppMemoryGroup` struct. As you iterate over the process list, you either find an existing entry in the map and update it, or create a new one. This is sometimes called the "lookup-or-create" pattern — it is identical to what you might do with a dictionary in Python or a Map in JavaScript.
- The reason the map value is a pointer (`*AppMemoryGroup`) rather than a plain `AppMemoryGroup` is important: when you retrieve a struct from a map in Go, you get a copy. If you modify the copy, the map still holds the original. By storing pointers, every lookup gives you direct access to the same struct in memory, so modifications (adding RSS, incrementing the count) take effect immediately without needing to write the value back into the map.
- This grouping step is the core feature that makes the app more useful than a raw `ps` listing. macOS Activity Monitor shows individual processes — you have to manually add them up to know how much total memory Chrome is using. Your app will do that automatically.

> **Go concept — maps:** A Go map is a hash table, similar to Python's `dict` or JavaScript's `Map`. You declare one with `make(map[KeyType]ValueType)`. You read from it with `value := m[key]` and write with `m[key] = value`. Unlike slices, maps must be initialized with `make` before you can write to them — writing to a nil map causes a runtime panic. Reading from a nil or empty map is safe and returns the zero value for the value type.

### 5.2 Write `groupByApp(processes []ProcessInfo) map[string]*AppMemoryGroup`

- Open `processes.go` — you will add this function below your existing `parseAllProcesses` function.
- Define an unexported function named `groupByApp`. It takes one parameter: `processes` of type `[]ProcessInfo`. It returns a `map[string]*AppMemoryGroup`.
- At the start of the function, create an empty map using `make(map[string]*AppMemoryGroup)`. Store it in a variable — this will be your accumulator.
- Iterate over the `processes` slice using a `for range` loop. On each iteration you have access to a single `ProcessInfo`. You need its `Command` field as the map key.
- For each process, look up `Command` in the map. Go maps support a two-value lookup: `entry, exists := m[key]`, where `exists` is a `bool` that is `true` if the key was found and `false` if it was not. This is the idiomatic way to check for key existence in Go — do not use a separate "contains" check.
- If the key does not exist (`exists` is false), create a new `AppMemoryGroup` struct with `Name` set to the process's `Command`, `MemoryKB` set to 0, and `ProcessCount` set to 0. Take its address with `&` to get a pointer, and store that pointer in the map under the `Command` key. Then assign the pointer to `entry` so the next steps work uniformly whether the entry was new or existing.
- Now that `entry` points to a valid `AppMemoryGroup` (either pre-existing or just created), add the process's `RSS` value to `entry.MemoryKB` and increment `entry.ProcessCount` by 1. Because `entry` is a pointer, these modifications affect the struct stored in the map directly.
- After the loop completes, return the map.
- This function needs no new imports — it uses only types already defined in your project.

### 5.3 Convert Map to Slice

- Your frontend expects a JSON array (which comes from a Go slice), not a JSON object (which would come from a Go map). You also cannot sort a map. So you need a function that extracts all the values from the map into a slice.
- In the same `processes.go` file, define an unexported function named `mapToSlice`. It takes one parameter: `grouped` of type `map[string]*AppMemoryGroup`. It returns a `[]AppMemoryGroup` — note this is a slice of values, not pointers. The frontend does not need pointers; it needs plain structs that will be serialized to JSON.
- Inside the function, declare a `[]AppMemoryGroup` variable to hold the results.
- Use a `for range` loop over the map. When you range over a map in Go, each iteration gives you a key and a value. You only need the value (the `*AppMemoryGroup` pointer). Use the blank identifier `_` for the key to tell Go you are intentionally ignoring it.
- On each iteration, dereference the pointer with `*` to get the `AppMemoryGroup` value, and append it to your result slice.
- Return the result slice after the loop.
- This function is intentionally simple — it does one thing (convert map values to a slice) and nothing else.

> **Go concept — map iteration order is random:** In Go, iterating over a map with `for range` visits keys in an unpredictable order that may change between runs. This is by design — the Go runtime intentionally randomizes map iteration order to prevent developers from depending on it. This means the slice returned by `mapToSlice` will have entries in an arbitrary order. That is fine for now — you will add sorting in Phase 6.

### 5.4 Wire It Together in MemoryService

- Open `memory.go`. Find your `GetMemoryUsage` method on `*MemoryService` — it currently returns a hardcoded slice of `AppMemoryGroup` structs from Phase 2.
- Delete the entire body of the method (the hardcoded slice literal). You will replace it with a four-step pipeline that calls the functions you have built across Phases 3, 4, and 5.
- Step 1: Call `getProcessList()`. This returns a `string` and an `error`. If the error is not nil, return an empty slice (`[]AppMemoryGroup{}`) and the error. This means you need to change the method signature to return two values: `([]AppMemoryGroup, error)`. Update the return type on the method declaration.
- Step 2: Pass the output string to `parseAllProcesses()`. This returns a `[]ProcessInfo`. No error to check — that function handles bad lines internally.
- Step 3: Pass the `[]ProcessInfo` to `groupByApp()`. This returns a `map[string]*AppMemoryGroup`.
- Step 4: Pass the map to `mapToSlice()`. This returns a `[]AppMemoryGroup`.
- Return the slice from step 4 and a `nil` error.
- The method body should now be about six to eight lines: one call per function, one error check after `getProcessList`, and a final return. Each function in the pipeline does exactly one transformation, and the method reads like a recipe: get raw output, parse it into structs, group by app name, convert to slice, return.
- Since you changed the return signature to include an `error`, Wails will handle this automatically. If your Go method returns `(SomeType, error)` and the error is non-nil, the generated JavaScript function will return a rejected Promise with the error message. Your frontend `catch` block in `App.svelte` already handles this (you wrote it in Phase 2).
- After making this change, run `wails3 generate bindings` from the project root to regenerate the JavaScript bindings. The bindings need to be regenerated because you changed the method's return signature. During `wails3 dev` this typically happens automatically, but running it manually ensures the bindings are fresh before you test.

### 5.5 Run & Verify in the App

- From the project root, run: `wails3 dev`
- The same frontend table from Phase 2 should appear — same three columns (Application, Memory KB, Processes) — but now populated with real data from your system instead of hardcoded values.
- No frontend changes are needed for this phase. The `AppMemoryGroup` struct has the same three fields with the same JSON tags as before. The `GetMemoryUsage` method has the same name and is on the same service. The only difference is that it now returns real data and can return an error. Since the generated JavaScript binding already returns a Promise and you already have a `catch` block, the contract between frontend and backend is unchanged.
- You should see dozens or hundreds of rows in the table. Look for applications you recognize: your web browser, Finder, WindowServer, Spotlight (mds_stores), and others.
- Open Activity Monitor (the macOS one, from Applications > Utilities) side by side with your app. Find a large application like your browser in both. Your app's memory number for that application should be roughly the sum of all the individual processes for that app shown in Activity Monitor. This is the whole point — your app groups what Activity Monitor shows individually.
- The rows will appear in a random order because Go maps do not preserve insertion order and you have not added sorting yet. Chrome might be in the middle of the list, tiny system daemons might be at the top. This is expected — sorting comes in Phase 6.
- If you see an empty table, open devtools (Cmd+Option+I) and check the Console for errors. The most likely issue is a compilation error in your Go code — check the terminal where `wails3 dev` is running for Go compiler output. Common mistakes at this stage: forgetting to change the return signature of `GetMemoryUsage` to include `error`, or a typo in one of the function names.
- If you see the table but with strange data (all zeros, or only one entry), add temporary `fmt.Println` statements in `GetMemoryUsage` to print intermediate results — the count of processes from `parseAllProcesses`, the number of keys in the map from `groupByApp`, and the length of the final slice. Run `go run .` (not `wails3 dev`) to see the output in your terminal, then remove the debug prints once you identify the issue.

**Checkpoint:** Real grouped memory data in the app. The "better than Activity Monitor" feature works. Data is unsorted and raw KB numbers, but it's real.

- The table displays real process data from your system. You see application names you recognize (your browser, Finder, WindowServer, etc.) with non-zero memory values and process counts.
- Applications that spawn multiple processes (like a web browser) show a `processCount` greater than 1, and their `memoryKB` value is the sum of all those processes. Compare this against macOS Activity Monitor to verify: find your browser, manually add up the memory of all its processes in Activity Monitor, and check that your app's total is in the same ballpark.
- The table has more than 20 rows (a typical macOS system has 100+ distinct application names after grouping). If you see fewer than 10, something is wrong with the parsing or grouping pipeline.
- Refreshing the app (close and reopen the window, or restart `wails3 dev`) shows slightly different memory numbers — proving the data is live, not cached or hardcoded.
- The rows appear in no particular order. Running the app multiple times may show the rows in different positions. This is correct and expected — sorting is added in Phase 6.

---

## Phase 6 — Sorting

### 6.1 Understand `sort.Slice`

- Go's standard library includes the `sort` package, which provides functions for sorting slices. The one you will use is `sort.Slice`. It takes two arguments: the slice you want to sort (of any type), and a "less" function that defines the ordering.
- The less function has the signature `func(i, j int) bool`. The parameters `i` and `j` are indices into the slice — not the elements themselves. Inside the function body, you use these indices to access the actual elements (for example, `result[i]` and `result[j]`), compare them however you want, and return `true` if the element at index `i` should come before the element at index `j` in the final sorted order. If you return `false`, element `j` comes first (or they stay in their current relative order if they are equal).
- This design is Go's way of sorting custom types without requiring them to implement an interface or define a comparison method on the type itself. You pass the comparison logic inline as an anonymous function (also called a closure). The anonymous function can reference variables from the surrounding scope — in this case, the slice you are sorting. This is similar to JavaScript's `Array.sort((a, b) => ...)` or Python's `sorted(key=...)`, but instead of receiving the elements directly, you receive their indices.
- `sort.Slice` sorts the slice in place — it modifies the original slice rather than returning a new one. There is no return value. This is important: you call it for its side effect, not for a result. After the call, the same slice variable holds the sorted data.

> **Go concept — anonymous functions (closures):** In Go, you can define a function without a name and pass it directly as an argument. The syntax is `func(params) returnType { body }` — it looks like a regular function definition but without a name, and it appears inline wherever a function value is expected. This anonymous function "closes over" variables from the enclosing scope, meaning it can read and modify them. In this case, the less function you pass to `sort.Slice` will reference your result slice by name — it does not receive the slice as a parameter, it simply uses it from the surrounding scope.

### 6.2 Sort by Memory Descending

- Open `memory.go` and find your `GetMemoryUsage` method on `*MemoryService`. Currently, the last two lines of the method body are the call to `mapToSlice` (which produces the final `[]AppMemoryGroup` slice) and the return statement.
- Add `"sort"` to the import block at the top of `memory.go`. This is the `sort` package from Go's standard library.
- Between the `mapToSlice` call and the return statement, add a call to `sort.Slice`. Pass it the result slice as the first argument. For the second argument, pass an anonymous function that takes two `int` parameters (conventionally named `i` and `j`) and returns a `bool`.
- Inside the anonymous function body, compare the `MemoryKB` field of the element at index `i` against the element at index `j`. You want descending order (highest memory first), so return `true` when the element at index `i` has a greater `MemoryKB` than the element at index `j`. Using the greater-than operator here flips the natural ascending order into descending order — the "heaviest" application will be sorted to the front of the slice.
- The anonymous function references the result slice variable by name from the enclosing scope. You do not pass the slice into the anonymous function — it captures it automatically because it is defined in the same scope. This is the closure behavior described in section 6.1.
- After this change, the method body should follow this sequence: get process list, parse all processes, group by app, convert map to slice, sort the slice by MemoryKB descending, return the sorted slice and nil error. The sort call adds exactly one statement — no new variables, no error handling needed.

### 6.3 Run & Verify

- From the project root, run: `wails3 dev`
- The table should now show applications sorted from highest memory usage to lowest. The first row should be the application consuming the most memory on your system — this is almost certainly your web browser (Chrome, Safari, Arc, or Firefox) or an Electron-based app like Slack or VS Code.
- Scroll down the table. You should see a clear descending trend: large applications with hundreds of thousands or millions of KB at the top, and tiny system daemons with single-digit KB values at the bottom.
- Open macOS Activity Monitor (Applications > Utilities > Activity Monitor). Click the "Memory" tab and sort by memory (click the "Memory" column header). Compare the top 5 applications in Activity Monitor against the top entries in your app. They should roughly agree on which applications are the biggest memory consumers, though your app's numbers will be higher for multi-process applications because you are summing all their processes together.
- Stop and restart `wails3 dev` a few times. The order should be consistent — the heaviest apps stay at the top. The exact memory values may shift slightly between runs (because processes allocate and free memory constantly), but the relative ranking should be stable.
- If the table is not sorted (rows appear in random order like before), make sure you added the `sort.Slice` call after `mapToSlice` and before the return. Also verify that you imported the `"sort"` package — Go will not compile if the import is missing, so if the app runs at all, the import is present. The most likely mistake is using less-than instead of greater-than in the comparison, which would give you ascending order (smallest first) instead of descending.

**Checkpoint:** Sorted by memory. The most interesting data is immediately visible.

---

## Phase 7 — Human-Readable Memory Formatting

### 7.1 Write `FormatMemory(kb int64) string`

- Open `memory.go`. You will add a standalone exported function (not a method on a struct) named `FormatMemory`. It takes one parameter `kb` of type `int64` and returns a `string`. It is exported (uppercase) because the bindings generator needs to see the types it touches, and keeping it exported also makes it easy to test later.
- Make sure `"fmt"` is in the import block at the top of `memory.go`. This is the only package this function needs.
- The function is a pure conversion with three thresholds. Check them from largest to smallest so the first match wins:
  - If `kb` is greater than or equal to `1_048_576` (which is 1024 * 1024, i.e. one gigabyte in KB), divide `kb` by `1_048_576.0` to get the value in gigabytes. The `.0` is important — it forces Go to perform floating-point division instead of integer division. Format the result with `fmt.Sprintf("%.1f GB", value)`. The `%.1f` verb formats a float with exactly one decimal place.
  - Otherwise, if `kb` is greater than or equal to `1024` (one megabyte in KB), divide `kb` by `1024.0` and format with `fmt.Sprintf("%.1f MB", value)`.
  - Otherwise (for values under 1024 KB), format with `fmt.Sprintf("%d KB", kb)`. The `%d` verb formats an integer with no decimal places — small values do not need fractional precision.
- Return the formatted string from each branch. The function should have three return paths — one per threshold.
- This function is a "pure function" — it has no side effects, depends only on its input, and always produces the same output for the same input. Pure functions are trivial to test and reason about.

> **Go concept — integer vs float division:** In Go, dividing two integers always produces an integer — `7 / 2` gives `3`, not `3.5`. To get a floating-point result, at least one operand must be a float. You can either cast the integer explicitly with `float64(kb)` and divide by a float literal, or simply use a float literal for the divisor (`1048576.0`). When you divide an `int64` by a float literal, Go automatically converts the integer to `float64` for the operation. This behavior is different from Python (where `/` always gives a float) and JavaScript (where all numbers are floats).

### 7.2 Add `MemoryFormatted` Field to the Struct

- Open `memory.go` and find your `AppMemoryGroup` struct. It currently has three fields: `Name`, `MemoryKB`, and `ProcessCount`.
- Add a fourth field named `MemoryFormatted` of type `string`, with the JSON struct tag `` `json:"memoryFormatted"` ``. Place it right after `MemoryKB` so the memory-related fields are grouped together.
- You now have two memory fields on the struct: `MemoryKB` (an `int64` holding the raw value in kilobytes) and `MemoryFormatted` (a `string` holding the human-readable version like "2.3 GB"). This is a deliberate design choice: keep the raw numeric value for sorting and computation, and add a pre-formatted string for display. If you only kept the formatted string, you could not sort by memory numerically — comparing "2.3 GB" and "850.0 MB" as strings would give wrong results.
- Now you need to populate `MemoryFormatted` before returning data to the frontend. Open the `GetMemoryUsage` method on `*MemoryService`. Find the spot after `mapToSlice` returns the `[]AppMemoryGroup` slice and before the `sort.Slice` call. Add a `for` loop that iterates over the slice using `range`. You need to modify each element in place, so use the index form of the loop: `for i := range result`. Inside the loop body, set `result[i].MemoryFormatted` to the return value of `FormatMemory(result[i].MemoryKB)`.
- Use the index (`result[i]`) to modify the element, not a loop variable copy. When you write `for i, group := range result`, the `group` variable is a copy of the element — modifying `group.MemoryFormatted` would change the copy and leave the slice untouched. Using `result[i]` modifies the actual element in the slice. This is a very common Go gotcha.
- After this change, regenerate bindings by running `wails3 generate bindings` from the project root. The bindings need to be updated because you added a new field to `AppMemoryGroup` — the generated JavaScript model class needs to include `memoryFormatted`.

> **Go concept — range copies values:** When you write `for _, item := range mySlice`, the `item` variable is a copy of each element, not a reference to it. Modifying `item` does not modify the slice. To modify elements in place, use the index: `for i := range mySlice` and then access `mySlice[i]`. This surprises developers coming from Python (where loop variables are references to the same objects) and JavaScript (where `for...of` gives you references to objects). In Go, structs are value types — assigning or passing them always copies.

### 7.3 Update the Frontend Table

- Open `frontend/src/App.svelte`. Replace its entire contents with the following Svelte component. This version uses the `memoryFormatted` field instead of raw `memoryKB`, adds a proper header row, and includes basic CSS styling:

```svelte
<script>
  import { onMount } from "svelte";
  import { GetMemoryUsage } from "../bindings/github.com/LeoManrique/activity-monitor/memoryservice";

  let groups = [];

  onMount(async () => {
    try {
      groups = await GetMemoryUsage();
    } catch (err) {
      console.error("Failed to fetch memory usage:", err);
    }
  });
</script>

<main>
  <h1>Memory Usage</h1>
  <table>
    <thead>
      <tr>
        <th class="col-name">Name</th>
        <th class="col-memory">Memory</th>
        <th class="col-procs">Processes</th>
      </tr>
    </thead>
    <tbody>
      {#each groups as group}
        <tr>
          <td class="col-name">{group.name}</td>
          <td class="col-memory">{group.memoryFormatted}</td>
          <td class="col-procs">{group.processCount}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</main>

<style>
  main {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
    padding: 1rem;
    max-width: 700px;
    margin: 0 auto;
  }

  h1 {
    font-size: 1.4rem;
    margin-bottom: 0.75rem;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.85rem;
  }

  thead th {
    text-align: left;
    border-bottom: 2px solid #ccc;
    padding: 0.4rem 0.75rem;
    font-weight: 600;
  }

  tbody td {
    padding: 0.3rem 0.75rem;
    border-bottom: 1px solid #eee;
  }

  .col-memory,
  .col-procs {
    font-family: "SF Mono", Menlo, Consolas, "Liberation Mono", monospace;
    text-align: right;
    white-space: nowrap;
  }

  .col-name {
    text-align: left;
  }

  tbody tr:hover {
    background-color: #f5f5f5;
  }
</style>
```

- Key changes from the Phase 2 version:
  - The memory column now displays `group.memoryFormatted` instead of `group.memoryKB`. This shows "2.3 GB" or "145.2 MB" instead of raw kilobyte numbers.
  - The header row uses "Name", "Memory", and "Processes" — clean labels that match a real application.
  - The Memory and Processes columns use a monospace font (`SF Mono` on macOS, falling back to `Menlo` and then `Consolas`). Monospace fonts make numbers align vertically so you can visually scan down the column and compare values. The body text uses the system UI font stack for readability.
  - The Memory and Processes columns are right-aligned with `text-align: right`. Right-aligning numeric data is a standard UI convention — it lines up the digits so that 2.3 GB and 145.2 MB are easy to compare at a glance. The Name column stays left-aligned.
  - A subtle row hover effect (`#f5f5f5` background) makes it easier to track which row you are reading across the columns.
- Remember: since you are running `wails3 dev`, saving this Svelte file triggers Vite hot-reload. The changes appear in the Wails window within a second or two without restarting the Go backend. If you changed Go code in this phase, the dev server recompiles Go automatically as well.

### 7.4 Run & Verify

- If `wails3 dev` is not already running, start it from the project root.
- The table should now display human-readable memory values: you should see entries like "2.3 GB", "145.2 MB", or "512 KB" in the Memory column instead of raw numbers like "2411724" or "148684".
- The heaviest application (usually your web browser) should show a value in the GB range. Mid-range applications like Slack, VS Code, or Xcode should be in the hundreds of MB. Small system daemons near the bottom of the list should show values in the low MB or KB range.
- Verify the table is still sorted by memory descending — the largest value should be at the top. Sorting still works correctly because it operates on `MemoryKB` (the raw `int64`), not on `MemoryFormatted` (the display string).
- Check that the header row is visible: "Name", "Memory", and "Processes" should appear above the data rows, separated by a horizontal line.
- Check that the Memory and Processes columns are right-aligned and rendered in a monospace font. The Name column should be left-aligned and in a regular proportional font.
- Open devtools (Cmd+Option+I), go to the Console tab, and inspect the data. If you can access the `groups` variable, verify that each object has both `memoryKB` (a number) and `memoryFormatted` (a string). Both fields travel from Go to the frontend — one for potential future use (sorting, calculations), one for display.
- Try a quick sanity check on the formatting thresholds: find an entry showing MB and verify it makes sense. For example, an entry showing "145.2 MB" should have a `memoryKB` value around 148,700 (because 145.2 * 1024 is approximately 148,685). If the numbers do not line up, double-check the divisors in your `FormatMemory` function — a common mistake is dividing by 1,000 instead of 1,024 or using the wrong threshold constant.

**Checkpoint:** Human-readable, sorted memory table. This is where it starts feeling like a real app.

- The Memory column displays formatted strings like "2.3 GB", "145.2 MB", or "512 KB" — not raw kilobyte integers.
- The formatting thresholds are correct: values of 1,048,576 KB and above show as GB, values of 1,024 KB and above show as MB, and smaller values show as KB. Each formatted value has one decimal place for GB and MB, and no decimal for KB.
- The table is still sorted by memory descending. The sort was not broken by adding the formatted field — the heaviest application is still in the first row.
- The header row ("Name | Memory | Processes") is visible and the columns are properly aligned: Name is left-aligned, Memory and Processes are right-aligned in a monospace font.
- Hovering over a table row highlights it with a subtle background color change, confirming the CSS is applied.

---

## Phase 8 — Auto-Refresh

### 8.1 Add `setInterval` Polling in JS

- No Go changes are needed for this phase. Everything happens in the Svelte frontend. The `GetMemoryUsage` method you built in previous phases already returns fresh data every time it is called — it runs the `ps` command, parses the output, groups, sorts, and returns. You just need to call it repeatedly.
- The browser API `setInterval(callback, milliseconds)` calls a function on a repeating timer. You will use it to call `GetMemoryUsage` every 3000 milliseconds (3 seconds). Three seconds is a good balance: frequent enough that the data feels live, infrequent enough that you are not hammering the system with `ps` commands.
- In Svelte, the right place to start a timer is inside the `onMount` lifecycle hook. You already have `onMount` from Phase 7 — you will expand it to set up the interval after the initial fetch. You also need `onDestroy` from Svelte to clean up the interval when the component is removed from the DOM. Without cleanup, the interval would keep firing even after the component is gone, causing memory leaks and errors in the console.
- Svelte's reactivity model makes this straightforward: when you reassign a variable declared with `let`, Svelte automatically re-renders any part of the template that references it. So `groups = await GetMemoryUsage()` inside the interval callback is all you need — Svelte sees the assignment, diffs the DOM internally, and updates only the table rows that changed. You do not need to manually clear and rebuild the table.
- Open `frontend/src/App.svelte` and replace its entire contents with the component below. The key differences from the Phase 7 version are: (1) a `refreshData` async function extracted so both the initial load and the interval can call it, (2) `setInterval` started inside `onMount`, (3) `onDestroy` added to clear the interval, and (4) a `lastUpdated` variable that tracks when data was last fetched.

```svelte
<script>
  import { onMount, onDestroy } from "svelte";
  import { GetMemoryUsage } from "../bindings/github.com/LeoManrique/activity-monitor/memoryservice";

  let groups = [];
  let lastUpdated = "";
  let intervalId;

  async function refreshData() {
    try {
      groups = await GetMemoryUsage();
      lastUpdated = new Date().toLocaleTimeString();
    } catch (err) {
      console.error("Failed to fetch memory usage:", err);
    }
  }

  onMount(() => {
    refreshData();
    intervalId = setInterval(refreshData, 3000);
  });

  onDestroy(() => {
    if (intervalId) {
      clearInterval(intervalId);
    }
  });
</script>

<main>
  <h1>Memory Usage</h1>
  <table>
    <thead>
      <tr>
        <th class="col-name">Name</th>
        <th class="col-memory">Memory</th>
        <th class="col-procs">Processes</th>
      </tr>
    </thead>
    <tbody>
      {#each groups as group}
        <tr>
          <td class="col-name">{group.name}</td>
          <td class="col-memory">{group.memoryFormatted}</td>
          <td class="col-procs">{group.processCount}</td>
        </tr>
      {/each}
    </tbody>
  </table>
  {#if lastUpdated}
    <span class="last-updated">Last updated: {lastUpdated}</span>
  {/if}
</main>

<style>
  main {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
    padding: 1rem;
    max-width: 700px;
    margin: 0 auto;
  }

  h1 {
    font-size: 1.4rem;
    margin-bottom: 0.75rem;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.85rem;
  }

  thead th {
    text-align: left;
    border-bottom: 2px solid #ccc;
    padding: 0.4rem 0.75rem;
    font-weight: 600;
  }

  tbody td {
    padding: 0.3rem 0.75rem;
    border-bottom: 1px solid #eee;
  }

  .col-memory,
  .col-procs {
    font-family: "SF Mono", Menlo, Consolas, "Liberation Mono", monospace;
    text-align: right;
    white-space: nowrap;
  }

  .col-name {
    text-align: left;
  }

  tbody tr:hover {
    background-color: #f5f5f5;
  }

  .last-updated {
    display: block;
    margin-top: 0.75rem;
    font-size: 0.75rem;
    color: #888;
  }
</style>
```

- Walk through the key pieces:
  - `refreshData` is an `async` function that calls `GetMemoryUsage`, assigns the result to `groups` (triggering Svelte's re-render), and updates `lastUpdated` with the current time. Extracting this into a named function avoids duplicating the fetch logic between the initial load and the interval callback.
  - Inside `onMount`, `refreshData()` is called immediately (so the table appears right away, not after a 3-second delay), and then `setInterval(refreshData, 3000)` schedules it to repeat every 3 seconds. The interval ID is stored in `intervalId` so it can be cleared later.
  - `onDestroy` is Svelte's cleanup lifecycle hook — it runs when the component is removed from the DOM. Calling `clearInterval(intervalId)` inside it stops the timer. The `if (intervalId)` guard is defensive — it prevents calling `clearInterval` with `undefined` if `onMount` never ran (which can happen in server-side rendering contexts, though Wails always runs client-side).
  - The `{#if lastUpdated}` block in the template only renders the timestamp span after the first successful fetch. Before that, `lastUpdated` is an empty string (falsy), so nothing appears. This avoids showing "Last updated: " with no time value during the brief moment before the first fetch completes.

### 8.2 Show Last Updated Timestamp

- The timestamp is already included in the component above. The `lastUpdated` variable is set inside `refreshData` using `new Date().toLocaleTimeString()`, which produces a locale-aware time string like "2:45:12 PM" (the exact format depends on your system's locale settings).
- The `<span class="last-updated">` sits below the table and displays the time of the most recent successful data fetch. Every 3 seconds, when `refreshData` runs and reassigns `lastUpdated`, Svelte detects the change and updates the span's text content automatically.
- This timestamp serves as a visual heartbeat — proof that the polling loop is alive and working. If the timestamp stops changing, you know the interval has stalled or `GetMemoryUsage` is hanging. During development, watch the seconds tick forward every 3 seconds to confirm the interval is firing.
- The CSS styles the timestamp in a small, muted gray font (`0.75rem`, `color: #888`) so it is visible but does not compete with the table data for attention. It sits below the table with a small top margin to separate it visually.

### 8.3 Run & Verify

- From the project root, run: `wails3 dev`
- The table should appear immediately with sorted memory data, just as it did in Phase 7. Below the table, you should see a "Last updated:" line with the current time.
- Watch the timestamp for 10-15 seconds. It should update every 3 seconds — you will see the seconds portion of the time advance in 3-second steps (for example, "2:45:12 PM", then "2:45:15 PM", then "2:45:18 PM"). This confirms the interval is running.
- Watch the memory values in the table. You should see numbers shift slightly every 3 seconds as processes allocate and free memory. The changes will be subtle — a few hundred KB here and there — but they prove the data is being refetched from Go each time, not cached.
- Now test that new applications appear: open a new application you do not currently have running — for example, open TextEdit or Preview from Spotlight. Within 3-6 seconds (one or two refresh cycles), a new row should appear in the table for that application.
- Test that closed applications disappear: quit the application you just opened (Cmd+Q, not just close the window). Within 3-6 seconds, its row should vanish from the table. This proves the entire pipeline runs fresh each cycle — `ps` is called, the output is re-parsed, re-grouped, re-sorted, and the frontend receives a completely new data set.
- Open devtools (Cmd+Option+I) and go to the Console tab. You should not see any errors appearing every 3 seconds. If you see "Failed to fetch memory usage" errors repeating, the Go backend is returning an error from `GetMemoryUsage` — check the terminal where `wails3 dev` is running for Go-side error output.
- To test the cleanup behavior: this is harder to observe directly, but you can verify it indirectly. In the devtools Console, check that there is only one interval running. If you were to hot-reload the Svelte component (by editing and saving `App.svelte`), the old component's `onDestroy` fires and clears the old interval, and the new component's `onMount` starts a fresh one. Without `onDestroy`, each hot-reload would add another interval, and you would see `GetMemoryUsage` being called 2, 3, 4 times per cycle — visible as duplicate network calls in the devtools Network tab or as multiple console logs if you add any.

**Checkpoint:** Live-updating memory monitor. The MVP is functionally complete.

---

## Phase 9 — Error Handling & Edge Cases

### 9.1 Handle `ps` Command Failure

- Open `processes.go` and review your `getProcessList` function. In Phase 3 you already wrote it to return `(string, error)`, and the caller in `GetMemoryUsage` (in `memory.go`) already checks for a non-nil error. Verify that this is the case: if `exec.Command("ps", ...).Output()` returns an error, your function should return an empty string and the error, and `GetMemoryUsage` should return an empty `[]AppMemoryGroup{}` and that same error. Do not call `log.Fatal` or `os.Exit` anywhere in this path — those would kill the entire desktop application. The error will travel through Wails to the frontend as a rejected Promise, and your existing `catch` block in the Svelte code will display it.
- Add the `"log"` package to your imports in `memory.go` if it is not already there. In the `GetMemoryUsage` method, just before you return the empty slice and error on `getProcessList` failure, add a call to `log.Println` that prints a message like "failed to run ps command:" followed by the error. This writes to the terminal where `wails3 dev` is running, giving you server-side visibility into failures without affecting the user-facing behavior. The frontend still receives the error via the rejected Promise — the log line is for your debugging convenience.
- To test this, temporarily change the command name inside `getProcessList` from `"ps"` to something that does not exist, like `"ps_broken"`. Run `wails3 dev`, open the app, and verify two things: (1) the terminal shows your log message with an error about the executable not being found, and (2) the frontend displays an error message instead of crashing or showing a blank screen. After confirming both, change the command name back to `"ps"`.

> **Go concept — `log.Println` vs `fmt.Println`:** Both print to standard output (actually `log` prints to standard error by default), but `log.Println` automatically prepends a timestamp to every message, which is valuable for debugging a long-running application. More importantly, `log` is safe to call from multiple goroutines simultaneously, while `fmt.Println` is not. For a desktop app where methods may be called concurrently from the frontend, prefer `log` for any diagnostic output.

### 9.2 Handle Malformed Lines

- Open `processes.go` and review your `parseAllProcesses` function from Phase 4. It should already skip lines where `parseProcessLine` returns a non-nil error — the loop calls `parseProcessLine`, checks for an error, and uses `continue` to move to the next line if parsing failed. Confirm this is exactly what your code does. If you wrote it differently (for example, if you return early on the first error, or if you collect errors into a slice), change it to the skip-and-continue pattern.
- Add a comment above the error check inside the loop explaining why you silently skip bad lines. Something like: "Some ps output lines have unexpected formats (kernel threads, zombie processes, or lines with missing fields). Skipping them is safe because they represent a tiny fraction of total memory and crashing on one bad line would make the entire refresh fail." This comment is for future-you or anyone else reading the code — it explains a design decision that might otherwise look like a bug.
- The key defensive checks in `parseProcessLine` should be: (1) that `strings.Fields` produces at least 3 elements, (2) that `strconv.Atoi` succeeds for the PID field, and (3) that `strconv.ParseInt` succeeds for the RSS field. If any of these fail, the function returns an error and `parseAllProcesses` skips that line. Walk through each check and confirm it is present. If you are missing any of these three, add it now.
- To verify your parser is resilient, add a temporary `fmt.Println` inside the error branch of the loop in `parseAllProcesses` that prints the line that failed and the error message. Run `wails3 dev` and check the terminal output. On a typical macOS system you may see zero to a few skipped lines — this is normal. Remove the debug print after inspecting the output.

### 9.3 Handle Empty Process Names

- Some macOS system processes report an empty command path, or a path that resolves to an empty string after `filepath.Base` processes it. For example, `filepath.Base("")` returns `"."`, and `filepath.Base("/")` returns `"/"`. Neither of those is a useful application name. Processes like these are typically low-level kernel tasks or zombie processes with minimal memory usage.
- Open `processes.go` and find the `parseProcessLine` function. After you call `filepath.Base` on the command string and before you store it in the `Command` field of the `ProcessInfo` struct, add a check: if the resulting name is empty, equals `"."`, or equals `"/"`, replace it with the string `"(unknown)"`. Use `strings.TrimSpace` on the name first to catch strings that are just whitespace.
- The reason to group these under `"(unknown)"` rather than skipping them entirely is visibility. If you skip them, their memory disappears from your totals and the user might wonder why the per-app numbers do not add up to the system total. Grouping them under a single `"(unknown)"` entry keeps the accounting honest — the user sees that some memory belongs to processes without a clear name, which is better than silently hiding it.
- After making this change, run `wails3 dev` and scroll through the table looking for an `"(unknown)"` entry. On most macOS systems you will find one with a small amount of memory and a handful of processes. If you do not see one, that is also fine — it means all processes on your system have valid command names. The important thing is that the code handles the case without panicking.

### 9.4 Consider Permissions

- When you run your app normally (not as root), the `ps` command cannot report full details for every process on the system. Some processes owned by root or other system users may show an RSS of 0, or may not appear at all in the output. This is a macOS security restriction, not a bug in your code.
- Open `processes.go` and add a comment at the top of the `getProcessList` function (below the function signature, above the first line of code) that explains this behavior. Write something like: "Note: when running without root privileges, some system processes may report 0 RSS or be omitted entirely by ps. This is expected macOS behavior. Running with sudo would show complete data but is not recommended for a desktop GUI application."
- Do not attempt to work around this limitation by running `ps` with `sudo` or by requesting elevated privileges. A desktop application should never require root access for basic monitoring. The data you get without root is accurate for all user-space applications (browsers, editors, chat apps) which are what users actually care about. The missing data is for low-level system daemons whose memory usage is typically small and stable.
- To see the difference for yourself, run `ps -axo pid,rss,comm | wc -l` in a normal terminal, then run `sudo ps -axo pid,rss,comm | wc -l`. The sudo version may show a few more processes. The difference is usually small — on a typical macOS system, the non-root version captures 95%+ of total process memory. This confirms that the limitation is minor and not worth complicating your app for.

**Checkpoint:** Robust error handling. The app gracefully handles edge cases without crashing.

- Temporarily break the `ps` command (change `"ps"` to `"ps_broken"` in `getProcessList`), run `wails3 dev`, and confirm: the terminal shows a log message with the error, and the frontend shows an error message (from your existing `catch` block) rather than crashing. Then revert the change.
- Review `parseAllProcesses` and confirm it contains a skip-and-continue pattern for lines where `parseProcessLine` returns an error. There is a comment explaining why bad lines are silently skipped.
- Review `parseProcessLine` and confirm it checks for at least 3 fields, validates PID with `strconv.Atoi`, and validates RSS with `strconv.ParseInt` — returning an error if any check fails.
- Review `parseProcessLine` and confirm that empty, `"."`, or `"/"` command names are replaced with `"(unknown)"`.
- Review `getProcessList` and confirm it has a comment explaining that non-root execution may cause some processes to report 0 RSS or be omitted.
- Run `wails3 dev` normally and confirm the app still works exactly as before — real data in the table, sorted by memory, auto-refreshing. The error handling changes are invisible during normal operation.

---

## Phase 10 — System Memory Summary

### 10.1 Get Total System Memory

- macOS exposes total physical RAM through the `sysctl` command. Run this in your terminal right now: `sysctl -n hw.memsize`. The `-n` flag suppresses the key name and prints only the value. You will see a single number like `17179869184` — this is your total RAM in bytes. On a 16 GB Mac this is 16 * 1024 * 1024 * 1024 = 17179869184.
- Open `processes.go`. You will add a new unexported function named `getTotalMemoryKB`. It takes no arguments and returns two values: an `int64` and an `error`.
- Inside the function, use `exec.Command("sysctl", "-n", "hw.memsize")` to create the command, then call `.Output()` to run it and capture stdout. Handle the error the same way you did in `getProcessList` — if non-nil, return 0 and the error.
- The output from `sysctl` is a byte slice containing the number followed by a trailing newline (for example, `"17179869184\n"`). Convert it to a string, then use `strings.TrimSpace` to strip the newline. Without trimming, `strconv.ParseInt` will fail because it does not accept whitespace or newline characters in the input string.
- Parse the trimmed string into an `int64` using `strconv.ParseInt` with base 10 and bit size 64. If parsing fails, return 0 and the error.
- The value is in bytes, but your app works in kilobytes (RSS from `ps` is in KB). Divide the parsed value by 1024 to convert bytes to kilobytes, and return the result with a `nil` error.
- An alternative approach would be using `/usr/sbin/system_profiler SPHardwareDataType`, which outputs human-readable hardware info including memory. However, `system_profiler` is dramatically slower (it takes 1-2 seconds to run versus milliseconds for `sysctl`) and requires parsing a multi-line formatted output. Stick with `sysctl` — it is purpose-built for reading single kernel parameters and returns instantly.

> **Go concept — `strings.TrimSpace`:** This function removes all leading and trailing whitespace characters (spaces, tabs, newlines, carriage returns) from a string and returns a new string. It is essential when processing command output because most Unix commands append a trailing newline to their stdout. If you forget to trim and try to parse `"17179869184\n"` as a number, you will get a parse error that is frustratingly hard to debug because the newline is invisible when printed.

### 10.2 Calculate Used vs Available

- You will approximate "used memory" by summing the RSS of all processes from your existing process list. This is the same data you already fetch and parse in `GetMemoryUsage` — you will reuse it rather than running `ps` a second time.
- This approach is an approximation, not an exact measurement, for three reasons: (1) RSS double-counts shared memory pages — when multiple processes share a library (like the system's C library), each process reports those shared pages in its own RSS, so the sum overstates actual physical usage. (2) The operating system uses memory for its own purposes (file caches, kernel buffers, GPU memory) that does not appear in any process's RSS. (3) Between the time `ps` runs and you sum the values, processes may have started, stopped, or changed their memory footprint.
- Despite these limitations, the RSS sum is a useful ballpark figure. It tells the user approximately how much memory their running applications are consuming, which is the question most people actually have. macOS Activity Monitor itself shows multiple memory categories ("App Memory", "Wired", "Compressed") precisely because there is no single "used" number — your approximation is honest and practical.
- "Available" memory is computed by subtracting the used total from the system total: `AvailableKB = TotalKB - UsedKB`. This is also approximate for the same reasons, but it gives the user a sense of headroom.
- Open `processes.go` and add a new unexported function named `sumAllRSS`. It takes one parameter: `processes` of type `[]ProcessInfo`. It returns an `int64`. Inside the function, iterate over the slice with a `for range` loop, accumulating each process's `RSS` field into a running total. Return the total after the loop. This function is intentionally simple — it does one thing (sum a field) and nothing else.

### 10.3 Add `GetSystemMemory()` Method to MemoryService

- Open `memory.go`. Below your `AppMemoryGroup` struct, define a new struct named `SystemMemory` with six fields:
  - `TotalKB` of type `int64` with JSON tag `` `json:"totalKB"` ``
  - `UsedKB` of type `int64` with JSON tag `` `json:"usedKB"` ``
  - `AvailableKB` of type `int64` with JSON tag `` `json:"availableKB"` ``
  - `TotalFormatted` of type `string` with JSON tag `` `json:"totalFormatted"` ``
  - `UsedFormatted` of type `string` with JSON tag `` `json:"usedFormatted"` ``
  - `AvailableFormatted` of type `string` with JSON tag `` `json:"availableFormatted"` ``
- The raw KB fields exist for any future computation the frontend might need (progress bars, percentage calculations). The formatted fields are for display — same dual-field pattern you used with `MemoryFormatted` on `AppMemoryGroup` in Phase 7.
- Below the struct definition, add a new method named `GetSystemMemory` on `*MemoryService`. It takes no arguments and returns two values: a `SystemMemory` and an `error`.
- Inside the method body, follow this sequence:
  - Call `getTotalMemoryKB()`. If the error is non-nil, return an empty `SystemMemory{}` and the error.
  - Call `getProcessList()`. If the error is non-nil, return an empty `SystemMemory{}` and the error.
  - Pass the process list output to `parseAllProcesses()` to get a `[]ProcessInfo`.
  - Pass the `[]ProcessInfo` to `sumAllRSS()` to get the used KB total.
  - Compute available KB by subtracting used from total. If the result is negative (which can happen because RSS over-counts shared pages), clamp it to 0 — do not return a negative available value, as that would confuse the user.
  - Build a `SystemMemory` struct, populating all six fields. Use your existing `FormatMemory` function for the three formatted fields — call it once for each of `TotalKB`, `UsedKB`, and `AvailableKB`.
  - Return the struct and a `nil` error.
- You do not need to register anything new in `main.go`. The method is on `MemoryService`, which is already registered with Wails via `application.NewService(&MemoryService{})`. In Wails 3, all exported methods on a registered service are automatically bound — adding a new method to an existing service makes it available to the frontend as soon as you regenerate bindings. There is no per-method registration step.
- From the project root, run: `wails3 generate bindings`. The output should now report 2 methods for `MemoryService` (previously it reported 1). Check `frontend/bindings/` — the `memoryservice` file should now contain both a `GetMemoryUsage` function and a `GetSystemMemory` function. You can also confirm by opening the generated file and looking for the new function name.

> **Go concept — multiple methods on one struct:** In Go, a struct can have as many methods as you want. You simply define additional functions with the same receiver type. There is no limit and no special syntax for adding the second, third, or tenth method — each is an independent function definition with the receiver specified. All methods on a struct have access to the struct's fields and can call each other. Wails discovers all exported methods via reflection when the service is registered, so adding `GetSystemMemory` alongside `GetMemoryUsage` on the same `*MemoryService` receiver is all you need.

### 10.4 Add a Summary Header in the Frontend

- Open `frontend/src/App.svelte`. Replace its entire contents with the following component. The key additions compared to Phase 8 are: (1) an import for `GetSystemMemory` alongside `GetMemoryUsage`, (2) a `systemMemory` variable to hold the system totals, (3) both methods called inside `refreshData`, and (4) a summary header div above the table displaying the system memory overview.

```svelte
<script>
  import { onMount, onDestroy } from "svelte";
  import { GetMemoryUsage } from "../bindings/github.com/LeoManrique/activity-monitor/memoryservice";
  import { GetSystemMemory } from "../bindings/github.com/LeoManrique/activity-monitor/memoryservice";

  let groups = [];
  let systemMemory = null;
  let lastUpdated = "";
  let intervalId;

  async function refreshData() {
    try {
      const [memResult, sysResult] = await Promise.all([
        GetMemoryUsage(),
        GetSystemMemory(),
      ]);
      groups = memResult;
      systemMemory = sysResult;
      lastUpdated = new Date().toLocaleTimeString();
    } catch (err) {
      console.error("Failed to fetch memory data:", err);
    }
  }

  onMount(() => {
    refreshData();
    intervalId = setInterval(refreshData, 3000);
  });

  onDestroy(() => {
    if (intervalId) {
      clearInterval(intervalId);
    }
  });
</script>

<main>
  <h1>Memory Usage</h1>

  {#if systemMemory}
    <div class="system-summary">
      <span class="summary-used">{systemMemory.usedFormatted} used</span>
      <span class="summary-sep">/</span>
      <span class="summary-total">{systemMemory.totalFormatted} total</span>
      <span class="summary-available">({systemMemory.availableFormatted} available)</span>
    </div>
  {/if}

  <table>
    <thead>
      <tr>
        <th class="col-name">Name</th>
        <th class="col-memory">Memory</th>
        <th class="col-procs">Processes</th>
      </tr>
    </thead>
    <tbody>
      {#each groups as group}
        <tr>
          <td class="col-name">{group.name}</td>
          <td class="col-memory">{group.memoryFormatted}</td>
          <td class="col-procs">{group.processCount}</td>
        </tr>
      {/each}
    </tbody>
  </table>

  {#if lastUpdated}
    <span class="last-updated">Last updated: {lastUpdated}</span>
  {/if}
</main>

<style>
  main {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
    padding: 1rem;
    max-width: 700px;
    margin: 0 auto;
  }

  h1 {
    font-size: 1.4rem;
    margin-bottom: 0.75rem;
  }

  .system-summary {
    background-color: #f0f4f8;
    border-radius: 6px;
    padding: 0.6rem 1rem;
    margin-bottom: 1rem;
    font-size: 0.95rem;
    font-family: "SF Mono", Menlo, Consolas, "Liberation Mono", monospace;
  }

  .summary-used {
    font-weight: 600;
  }

  .summary-sep {
    color: #999;
    margin: 0 0.3rem;
  }

  .summary-total {
    color: #555;
  }

  .summary-available {
    color: #888;
    margin-left: 0.4rem;
    font-size: 0.85rem;
  }

  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.85rem;
  }

  thead th {
    text-align: left;
    border-bottom: 2px solid #ccc;
    padding: 0.4rem 0.75rem;
    font-weight: 600;
  }

  tbody td {
    padding: 0.3rem 0.75rem;
    border-bottom: 1px solid #eee;
  }

  .col-memory,
  .col-procs {
    font-family: "SF Mono", Menlo, Consolas, "Liberation Mono", monospace;
    text-align: right;
    white-space: nowrap;
  }

  .col-name {
    text-align: left;
  }

  tbody tr:hover {
    background-color: #f5f5f5;
  }

  .last-updated {
    display: block;
    margin-top: 0.75rem;
    font-size: 0.75rem;
    color: #888;
  }
</style>
```

- Walk through the key changes:
  - `GetSystemMemory` is imported from the same binding module as `GetMemoryUsage`. Both functions live in the same generated file because they are methods on the same Go service.
  - Inside `refreshData`, `Promise.all` calls both Go methods in parallel. This is important for performance — each method internally runs `ps` and does work, so running them concurrently saves time compared to awaiting them sequentially. The destructured result `[memResult, sysResult]` captures both return values.
  - The `systemMemory` variable starts as `null`. The `{#if systemMemory}` block in the template prevents rendering the summary header until the first fetch completes, avoiding a flash of empty content.
  - The summary header uses `systemMemory.usedFormatted`, `systemMemory.totalFormatted`, and `systemMemory.availableFormatted` — the pre-formatted strings from your Go struct. The property names match the JSON tags you defined on `SystemMemory`.
  - The summary div sits between the `<h1>` and the `<table>`, visually bridging the title and the per-app breakdown. It uses a light blue-gray background (`#f0f4f8`) and rounded corners to make it look like an info card.
- Remember: since this is a frontend-only change (Svelte/CSS), it hot-reloads instantly via Vite if `wails3 dev` is running. However, since you also made Go changes in sections 10.1-10.3, you need to restart `wails3 dev` (or let it auto-recompile) for the Go side to take effect. A good rule of thumb: **frontend changes hot-reload, Go changes need a recompile** (which `wails3 dev` handles automatically when it detects Go file changes, but it takes a few seconds).

### 10.5 Run & Verify

- From the project root, run: `wails3 dev`
- Above the table, you should see the system memory summary header displaying something like "12.4 GB used / 16.0 GB total (3.6 GB available)". The exact numbers depend on your Mac's RAM and what you have running.
- The "total" value should match your physical RAM. If you have a 16 GB Mac, you should see "16.0 GB total". If you have 8 GB, "8.0 GB total". Verify this matches what Apple menu > About This Mac reports for memory.
- The "used" value will be an approximation and will likely be higher than what Activity Monitor shows as "Memory Used" in its Memory tab. This is expected — your number is the sum of all RSS values, which double-counts shared pages. Activity Monitor has access to private kernel APIs that give it a more accurate breakdown. Your number is a useful upper bound, not an exact measurement.
- The table below the header should still display per-app memory data, sorted by memory descending, with formatted values and auto-refresh — all the behavior from previous phases is preserved.
- Watch both the summary header and the table for 10-15 seconds. Both should update every 3 seconds (matching the interval timer). The "used" and "available" values in the header should shift slightly as processes allocate and free memory. The "total" value should never change — physical RAM is fixed.
- Open macOS Activity Monitor (Applications > Utilities > Activity Monitor), click the Memory tab, and compare: your "total" should match Activity Monitor's "Physical Memory" exactly. Your "used" will be higher than Activity Monitor's "Memory Used" for the reasons explained above. The key thing is that the values are in the same ballpark — if Activity Monitor shows 10 GB used and your app shows 14 GB, that is reasonable (the difference is shared page double-counting). If your app shows 2 GB or 200 GB, something is wrong with the calculation.
- Open devtools (Cmd+Option+I), go to the Console tab. You should not see any new errors. If you see "Failed to fetch memory data" errors, check the terminal where `wails3 dev` is running — the Go side may be returning an error from `getTotalMemoryKB` (perhaps `sysctl` failed) or from `getProcessList`.
- To verify the clamping logic: this is hard to trigger naturally, but you can test it by temporarily making `sumAllRSS` return an artificially large number (larger than total RAM). The available value in the header should show "0 KB" rather than a negative number. Revert the change after testing.

**Checkpoint:** System-level overview + per-app detail. The app tells a complete memory story.

- The summary header appears above the table and displays three values: used, total, and available — all in human-readable format (e.g., "12.4 GB used / 16.0 GB total (3.6 GB available)").
- The "total" value matches your Mac's physical RAM exactly. Compare it against Apple menu > About This Mac. A 16 GB Mac should show "16.0 GB total", an 8 GB Mac should show "8.0 GB total".
- The "used" value is a positive number smaller than (or in edge cases roughly equal to) the total. It is likely higher than what Activity Monitor reports as "Memory Used" — this is expected due to RSS double-counting shared pages.
- The "available" value is never negative. It equals total minus used, or "0 KB" if the RSS sum exceeds total RAM.
- Both the summary header and the table update every 3 seconds. The "used" and "available" numbers shift slightly between refreshes as processes allocate and free memory. The "total" number never changes.
- The per-app table still works exactly as before: sorted by memory descending, formatted values, hover highlighting, last-updated timestamp. No Phase 8 or Phase 9 behavior was broken.
- Run `sysctl -n hw.memsize` in your terminal, divide the result by 1024, and verify that your app's `totalKB` value (visible in devtools by inspecting the `systemMemory` object) matches this number exactly.

---

## Phase 11 — Build & Package

### 11.1 Production Build

- Stop `wails3 dev` if it is running (Ctrl+C). A production build is a separate step — you do not need the dev server running.
- From the project root, run: `wails3 build`
- This command does three things in sequence: (1) it builds your frontend by running `npm run build` inside `frontend/`, which produces optimized, minified HTML/CSS/JS in `frontend/dist/`; (2) it compiles all your Go code (including your `MemoryService`, `FormatMemory`, and everything in `main.go`); (3) it uses Go's `embed` directive to bake the contents of `frontend/dist/` directly into the compiled binary. The result is a single standalone executable that contains everything — no external files, no Node.js runtime, no separate frontend folder needed.
- The compiled binary is placed in the `bin/` directory at your project root. On macOS, the output file will be named after your project (e.g., `activity-monitor`). Check it exists by running `ls -la bin/`.
- To go one step further and produce a proper macOS `.app` bundle (the kind you can double-click in Finder and drag to Applications), run: `wails3 package`
- The `wails3 package` command wraps your compiled binary into a standard macOS application bundle structure. The output is `bin/activity-monitor.app` (or whatever your project name is). Inside the `.app` bundle, the `Contents/MacOS/` directory holds the actual executable, `Contents/Info.plist` holds application metadata, and `Contents/Resources/` holds the app icon. You can inspect this structure by running `ls -R bin/activity-monitor.app/Contents/` in your terminal.
- The entire `.app` bundle is self-contained. You can copy it to another Mac and it will run without installing Go, Node.js, or any dependencies. This is the key advantage of the Wails build approach — everything is compiled and embedded into a single deliverable.

> **Go concept — `//go:embed`:** The `embed` package, introduced in Go 1.16, lets you include files and directories inside your compiled binary at build time. In your `main.go`, the directive `//go:embed frontend/dist` tells the Go compiler to read every file under `frontend/dist/` and store them as byte data inside the binary. At runtime, Wails serves these embedded files to the webview instead of reading from disk. This is why the production app does not need a `frontend/` folder next to it — the assets are literally inside the executable.

### 11.2 Test the Built App

- Open Finder and navigate to your project's `bin/` directory (inside `activity-monitor/bin/`). You should see `activity-monitor.app` (if you ran `wails3 package`) or just the raw `activity-monitor` binary (if you only ran `wails3 build`).
- Double-click `activity-monitor.app` to launch it. A native macOS window should appear showing your memory usage table — the same UI you saw during development, but now running as a standalone application with no terminal, no dev server, and no Vite process.
- Verify the full feature set works: the table should display real process data (not hardcoded), sorted by memory descending, with human-readable formatted values. The system memory summary header should appear above the table. The data should auto-refresh every 3 seconds — watch the "Last updated" timestamp to confirm.
- If macOS blocks the app with a "cannot be opened because the developer cannot be verified" dialog, go to System Settings > Privacy & Security, scroll down, and click "Open Anyway" next to the message about your app. This happens because the app is not code-signed with an Apple Developer certificate — it is expected for local development builds.
- You can also launch the built app from the terminal to see any Go-side log output. Run `./bin/activity-monitor.app/Contents/MacOS/activity-monitor` directly. Any `fmt.Println` or `log.Println` output from your Go code will appear in the terminal. This is useful for debugging issues that only appear in the production build.
- Try quitting and relaunching the app a few times to confirm it starts reliably. Quit with Cmd+Q or by right-clicking the dock icon and selecting Quit.

### 11.3 Debugging Cheatsheet

- **"binding not found" or "GetMemoryUsage is not a function" in the browser console:** The frontend is trying to call a Go method that Wails does not know about. This usually means you forgot to run `wails3 generate bindings` after changing a Go method signature, or the method name is unexported (starts with a lowercase letter). Fix: verify the method is exported (uppercase first letter), re-run `wails3 generate bindings`, and check that the import path in your Svelte file matches the actual folder names inside `frontend/bindings/`.
- **Nil pointer panic (the app crashes with a stack trace mentioning "nil pointer dereference"):** You are calling a method on a struct that was never initialized, or accessing a field on a nil value. The most common cause in this project is forgetting the `&` when registering your service — `application.NewService(MemoryService{})` instead of `application.NewService(&MemoryService{})`. Another cause: your `getProcessList` function returned `nil` and downstream code tried to iterate or index into it without checking. Fix: read the stack trace — it tells you the exact file and line number. Go stack traces are verbose but precise.
- **"permission denied" when running `ps` or the built app:** On macOS, some process information may be restricted. If `ps` fails with a permission error, open System Settings > Privacy & Security > Full Disk Access and add your terminal application (Terminal.app or iTerm). For the built `.app`, you may need to add it there as well. You can test whether `ps` works outside the app by running `ps -eo pid,rss,comm` directly in your terminal and checking for errors.
- **Checking `ps` output directly in the terminal:** If the app shows unexpected data (wrong process names, zero memory values, missing processes), run the same `ps` command your Go code uses directly in your terminal: `ps -eo pid,rss,comm`. Compare the raw output against what your app displays. This isolates whether the problem is in the `ps` output itself or in your Go parsing logic. Pipe it through `sort` to compare ordering: `ps -eo pid,rss,comm | sort -k2 -rn | head -20`.
- **Using `fmt.Println` and `log.Println` for Go debugging:** During development with `wails3 dev`, any `fmt.Println(...)` or `log.Println(...)` calls in your Go code print to the terminal where `wails3 dev` is running. Use these liberally when debugging — print the raw `ps` output before parsing, print the length of your grouped slice, print individual field values. The difference between the two: `log.Println` includes a timestamp prefix, `fmt.Println` does not. Neither one appears in the Wails window — they only go to the terminal (stdout/stderr).
- **Using browser devtools in the Wails window:** Press Cmd+Option+I (or right-click > Inspect Element) to open devtools during development. Use the Console tab to check for JavaScript errors, the Network tab to see binding calls, and the Elements tab to inspect the DOM. Note: devtools are available in `wails3 dev` mode. In a production build created with `wails3 build`, devtools are disabled by default. If you need devtools in a production build for testing, you can temporarily add the `-debug` flag: `wails3 build -debug`.
- **The app builds but shows a blank white window:** This usually means the frontend assets were not built or embedded correctly. Check that `frontend/dist/` exists and contains `index.html` by running `ls frontend/dist/`. If it is empty or missing, run `cd frontend && npm run build && cd ..` manually, then run `wails3 build` again. Another cause: a JavaScript error is crashing the app on startup — launch from the terminal (as described in 11.2) and look for errors, or rebuild with `-debug` and open devtools.
- **The table shows stale or empty data after a Go change:** When using `wails3 dev`, Go changes trigger an automatic recompile, but the frontend may still be using a cached version of the bindings. Stop `wails3 dev`, run `wails3 generate bindings` manually, then restart `wails3 dev`. This forces the bindings to regenerate from scratch.

**Checkpoint:** You have a standalone macOS app you built from scratch.

- Running `wails3 build` completes without errors and produces a binary in the `bin/` directory.
- Running `wails3 package` produces a `bin/activity-monitor.app` bundle that you can see in Finder.
- Double-clicking the `.app` in Finder launches a native window showing the memory usage table with real data — no terminal, no dev server required.
- The production app displays all features from previous phases: real process data, sorted by memory descending, human-readable memory formatting, system memory summary header, and auto-refresh every 3 seconds.
- You can launch the app from the terminal via `./bin/activity-monitor.app/Contents/MacOS/activity-monitor` and see any `fmt.Println`/`log.Println` output printed to the terminal.
- After quitting (Cmd+Q) and relaunching, the app starts reliably and shows fresh data.

---

## Phase 12 — Next Steps (Post-MVP)

### 12.1 Add CPU Usage Column

- Change the `ps` flags in `getProcessList` from `"pid,rss,comm"` to `"pid,rss,%cpu,comm"`. The `%cpu` column outputs a floating-point percentage (e.g., `12.3`). Run `ps -axo pid,rss,%cpu,comm` in your terminal first to see the new column position.
- Add a `CPU` field of type `float64` to `ProcessInfo`. In `parseProcessLine`, parse the new third field (index 2) using `strconv.ParseFloat` with bit size 64, and shift the command rejoining logic to start from index 3 instead of index 2.
- Add a `CPUPercent` field of type `float64` with JSON tag `` `json:"cpuPercent"` `` to `AppMemoryGroup`. In `groupByApp`, sum the CPU values the same way you sum RSS. In the frontend, add a fourth table column header "CPU" and display `group.cpuPercent` formatted with `toFixed(1)` and a `%` suffix.
- **Checkpoint:** The table shows a CPU % column per app group. Values update every 3 seconds. Chrome or similar apps should show non-zero values. Running `ps -axo pid,rss,%cpu,comm` in your terminal should produce numbers in the same ballpark as what the app displays.

### 12.2 Search / Filter Bar

- This is a frontend-only change — no Go code needed. Add a `let searchTerm = ""` variable in `App.svelte` and an `<input>` element above the table bound to it with `bind:value={searchTerm}`. Give it a placeholder like "Filter by name...".
- Create a reactive filtered list using `$:` syntax: `$: filteredGroups = groups.filter(g => g.name.toLowerCase().includes(searchTerm.toLowerCase()))`. Replace the `{#each groups as group}` block with `{#each filteredGroups as group}` so the table only shows matching rows.
- Style the input with the same font family and sizing as the table. Add `type="search"` so macOS renders a native clear button (the small "x") inside the input when text is present.
- **Checkpoint:** Typing "chrome" into the filter input instantly narrows the table to rows whose name contains "chrome" (case-insensitive). Clearing the input restores the full list. The filter survives a data refresh — if you type a filter and wait 3 seconds, the refreshed data is still filtered.

### 12.3 Click to Kill Process

- This requires a new Go method and a frontend change. You need process PIDs available in the frontend, so first add a `PIDs` field of type `[]int` with JSON tag `` `json:"pids"` `` to `AppMemoryGroup`. Populate it in `groupByApp` by appending each process's PID to the slice for its group.
- Add a new exported method `KillProcess` on `*MemoryService` that takes a single `int` parameter (the PID) and returns an `error`. Inside, use `os.FindProcess(pid)` from the `os` package to get an `*os.Process`, then call `.Kill()` on it. Return the error from `.Kill()`. You will need to import the `"os"` package.
- In the frontend, add a small "Kill" button in each table row. When clicked, show a `window.confirm()` dialog asking the user to confirm ("Kill processName?"). Only if confirmed, call the generated `KillProcess` binding with the first PID from the group's `pids` array. After the call resolves, trigger a `refreshData()` to update the table.
- **Checkpoint:** Clicking "Kill" on a row shows a browser confirmation dialog. Confirming it terminates the process (verify by checking Activity Monitor or running `ps` in terminal — the PID should be gone). The table refreshes and the killed process disappears or its memory drops. Canceling the dialog does nothing.

### 12.4 Memory Usage Bar/Chart

- This is a frontend-only enhancement. You already have `systemMemory.totalKB` and each group's `memoryKB` available in the frontend, so you can compute a percentage: `(group.memoryKB / systemMemory.totalKB) * 100`.
- Add a new `<td>` in each table row containing a `<div>` with a fixed height (e.g., 12px), a colored background (e.g., `#4a90d9`), `border-radius` for rounded corners, and a dynamic `width` set via inline style: `style="width: {percentage}%"`. Wrap it in a container `<div>` with a light gray background so the bar has a visible track.
- Cap the width at 100% using `Math.min(percentage, 100)` to handle the edge case where RSS double-counting makes a single app appear to exceed total RAM. Consider using a logarithmic scale or capping at the largest app's value instead of total RAM for better visual differentiation between smaller apps.
- **Checkpoint:** Each row shows a colored horizontal bar whose width is proportional to that app's share of total memory. Chrome or the largest memory consumer has the widest bar. Smaller apps have visibly narrower bars. Bars update every 3 seconds along with the data. No bar overflows its container.

### 12.5 Replace `ps` with `syscall`/`cgo`

- This is a significant rewrite of `getProcessList` and `parseAllProcesses` — replacing the `exec.Command("ps", ...)` shell-out approach with direct macOS kernel calls. It is considerably harder than any previous phase but teaches you `cgo`, `unsafe` pointers, and how operating systems expose process data.
- Use `syscall.Sysctl("kern.proc.all")` or the lower-level `sysctl` syscall with MIB `{CTL_KERN, KERN_PROC, KERN_PROC_ALL}` to get a raw byte buffer containing every process's `kinfo_proc` struct. You will need to import `"unsafe"` and `"syscall"` and cast the buffer to a slice of C-compatible structs. Look up the `kinfo_proc` struct layout in Apple's `<sys/sysctl.h>` header.
- For per-process memory (RSS), use `proc_pidinfo` with the `PROC_PIDTASKINFO` flavor via cgo. This requires writing a small C bridge: add `// #include <libproc.h>` and `import "C"` at the top of a new file (e.g., `procinfo_darwin.go`), then call `C.proc_pidinfo(C.int(pid), C.PROC_PIDTASKINFO, ...)` to fill a `proc_taskinfo` struct whose `pti_resident_size` field gives you RSS in bytes.
- Start small: replace just the process listing first (using `sysctl`), keeping your existing grouping and formatting logic. Once that works, replace the RSS reading with `proc_pidinfo`. Test at each step by comparing output against the `ps`-based version.
- **Checkpoint:** The app displays the same data as before (same app names, similar memory values, same grouping) but the Go code no longer calls `exec.Command("ps", ...)`. Running `wails3 dev` with the new implementation compiles without errors. Memory values are within 5-10% of what the old `ps`-based version showed (small differences are expected because `proc_pidinfo` reports bytes and timing differs).
