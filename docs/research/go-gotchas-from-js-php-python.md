# Go gotchas for programmers from JS/TS, PHP, and Python (and Go’s memory model)

Research notes from **primary sources only** (go.dev docs, language spec, go.dev/blog, pkg.go.dev / stdlib sources, official Go wiki on go.dev).  
Retrieved / verified: 2026-08-12.

Audience: engineers joining this OAuth / Go codebase from JS/TS, PHP, or Python. This is an onboarding research note, not a tutorial novel. Version-sensitive facts (especially **Go 1.22 loop scoping**, **GOGC / GOMEMLIMIT**, and **`omitempty` / `omitzero`**) are called out explicitly.

---

## Part A — Gotchas when coming from JS/TS, PHP, Python

### 1. Errors as values (not exceptions); panic / recover role

Go’s conventional failure path is a returned `error` value (the predeclared `error` interface). Callers check `if err != nil`. This is deliberate: multi-value returns report errors without overloading the primary result, and libraries encourage explicit checks at the call site rather than try/catch-style unwinding for ordinary failure.

Sources:

- [https://go.dev/blog/error-handling-and-go](https://go.dev/blog/error-handling-and-go)  
- [https://go.dev/doc/effective_go#errors](https://go.dev/doc/effective_go#errors)  
- Spec: [https://go.dev/ref/spec#Errors](https://go.dev/ref/spec#Errors)  
- FAQ (vs exceptions): [https://go.dev/doc/faq#exceptions](https://go.dev/doc/faq#exceptions)

**`panic` / `recover`** are *not* the everyday substitute for exceptions:

| Mechanism | Role |
| --- | --- |
| `error` return | Normal, expected failure (I/O, validation, auth) |
| `panic` | Stops the current goroutine’s ordinary flow; runs deferred functions while unwinding; ends the program if unrecovered |
| `recover` | Only useful **inside a deferred function**; stops panicking and returns the panic value |

Library convention: even when a package uses `panic` internally, its **exported API still presents `error` values**. Prefer `panic` for truly exceptional / “impossible” situations (or during init when setup cannot proceed).

Sources: [https://go.dev/blog/defer-panic-and-recover](https://go.dev/blog/defer-panic-and-recover), [https://go.dev/doc/effective_go#panic](https://go.dev/doc/effective_go#panic), [https://go.dev/ref/spec#Handling_panics](https://go.dev/ref/spec#Handling_panics)

---

### 2. Zero values

Uninitialized variables and newly allocated storage get a **zero value**: `false`, `0`, `""`, or `nil` for pointers, functions, interfaces, slices, channels, and maps. Zeroing is recursive (e.g. struct fields).

Sources: [https://go.dev/ref/spec#The_zero_value](https://go.dev/ref/spec#The_zero_value), [https://go.dev/doc/effective_go#zero](https://go.dev/doc/effective_go#zero)

Idiom: design types so the zero value is useful (`bytes.Buffer`, unlocked `sync.Mutex`).  
Source: [https://go.dev/doc/effective_go#allocation_new](https://go.dev/doc/effective_go#allocation_new)

Contrast with JS (`undefined`), PHP unset / `null`, Python `None`: Go has **no uninitialized-memory garbage**; you always get a defined zero.

---

### 3. Exported vs unexported (capitalization)

An identifier is **exported** (visible outside its package) iff:

1. Its first character is a Unicode uppercase letter (category Lu), and  
2. It is declared in the package block, or is a field / method name.

All other identifiers are unexported (package-private). There is no `public` / `private` keyword.

Source: [https://go.dev/ref/spec#Exported_identifiers](https://go.dev/ref/spec#Exported_identifiers)

Practical hit for JSON / reflection: only **exported** struct fields are visible to `encoding/json` and settable via reflection.  
Sources: [https://go.dev/blog/json](https://go.dev/blog/json), [https://go.dev/blog/laws-of-reflection](https://go.dev/blog/laws-of-reflection)

---

### 4. Pass-by-value; pointers vs values; method receivers

**Everything is passed by value** (a copy of the argument, as if assigned to the parameter). Passing a pointer copies the pointer, not the pointee. Map and slice values are descriptors that contain references to underlying data—copying the descriptor does not copy the data.

Sources: [https://go.dev/doc/faq#pass_by_value](https://go.dev/doc/faq#pass_by_value), [https://go.dev/ref/spec#Representation_of_values](https://go.dev/ref/spec#Representation_of_values)

**Method receivers** — choose pointer vs value deliberately:

| Need | Prefer |
| --- | --- |
| Method mutates receiver | Pointer receiver |
| Receiver is large | Pointer receiver (efficiency) |
| Consistency of method set | If some methods need pointers, often make all pointer receivers |
| Small / immutable semantics | Value receiver is fine |

Sources: [https://go.dev/doc/faq#methods_on_values_or_pointers](https://go.dev/doc/faq#methods_on_values_or_pointers), [https://go.dev/doc/effective_go#pointers_vs_values](https://go.dev/doc/effective_go#pointers_vs_values)

Unlike JS objects / PHP objects / Python objects (reference semantics by default for instances), a Go `struct` value is a full copy unless you use pointers (or reference-like headers: slice/map/chan).

---

### 5. Slices (headers, backing arrays, append, nil vs empty)

A slice is a **descriptor**: pointer to an underlying array, length, and capacity. Slicing does **not** copy element data; multiple slices can alias the same array. Modifying elements through one slice is visible through others that share storage.

Sources: [https://go.dev/blog/slices-intro](https://go.dev/blog/slices-intro), [https://go.dev/ref/spec#Slice_types](https://go.dev/ref/spec#Slice_types)

| Fact | Detail |
| --- | --- |
| Zero value | `nil`; `len` and `cap` are 0 |
| `append` on nil | Works; nil behaves like a zero-length slice for append |
| `append` growth | May allocate a new backing array; **always assign the result** (`s = append(s, x)`) |
| Nil vs empty | `var s []T` is nil; `s := []T{}` or `make([]T, 0)` is non-nil empty — both have length 0, but they are not the same value |

Sources: [https://go.dev/blog/slices-intro](https://go.dev/blog/slices-intro), [https://go.dev/ref/spec#Making_slices_maps_and_channels](https://go.dev/ref/spec#Making_slices_maps_and_channels), [https://go.dev/ref/spec#Composite_literals](https://go.dev/ref/spec#Composite_literals) (nil vs empty note)

**Gotcha — retaining large backing arrays:** a small subslice can keep an entire large array alive for the GC. Copy out the needed bytes if you must retain only a small piece.

Source: [https://go.dev/blog/slices-intro](https://go.dev/blog/slices-intro) (“A possible gotcha”)

---

### 6. Maps (nil write panic, missing key zero value, not concurrent-safe, iteration order)

| Behavior | Rule |
| --- | --- |
| Zero / nil map | Reading is fine (missing → element zero value); **writing panics** |
| Missing key | Indexing returns the **zero value** of the element type |
| Presence test | `v, ok := m[k]` (“comma ok”) |
| Concurrency | **Not safe** for concurrent read+write without synchronization |
| Iteration order | **Not specified**; may change from one `range` to the next |

Sources:

- [https://go.dev/blog/go-maps-in-action](https://go.dev/blog/go-maps-in-action)  
- Spec map types / range: [https://go.dev/ref/spec#Map_types](https://go.dev/ref/spec#Map_types), [https://go.dev/ref/spec#For_statements](https://go.dev/ref/spec#For_statements)  
- Race detector example (concurrent map): [https://go.dev/doc/articles/race_detector](https://go.dev/doc/articles/race_detector)

Contrast: Python `dict` / JS objects raise or yield `undefined` differently; PHP arrays are more forgiving. In Go, **`make` (or a literal) before first write**.

---

### 7. Strings (bytes vs runes, immutability, `len`)

- A string is effectively a **read-only** sequence of bytes (immutable).  
- Content may be arbitrary bytes; it is **not** required to be UTF-8 (though string *literals* without byte escapes are UTF-8).  
- Indexing / `len(s)` operate on **bytes**, not “characters”.  
- `for range` over a string iterates **Unicode code points (runes)**, with the index in **byte** offsets.  
- `rune` is an alias for `int32` (code point).

Sources: [https://go.dev/blog/strings](https://go.dev/blog/strings), [https://pkg.go.dev/builtin#string](https://pkg.go.dev/builtin#string)

Contrast: JS strings are UTF-16 code units; PHP strings are byte strings; Python 3 `str` is Unicode text. In Go, decide explicitly whether you mean bytes or runes (`unicode/utf8`, `[]rune` conversions).

---

### 8. Interfaces and typed-nil vs nil

An interface value holds a **(dynamic type, dynamic value)** pair. It is `nil` only when **both** are unset. Storing a nil **concrete** pointer inside an interface yields a **non-nil** interface:

```text
var p *MyError = nil
var err error = p   // err != nil  (type=*MyError, value=nil)
```

Classic bug: returning a nil concrete error pointer as `error` makes callers think there was always an error. Return an explicit untyped/`error` `nil` instead.

Sources: [https://go.dev/doc/faq#nil_error](https://go.dev/doc/faq#nil_error), [https://go.dev/blog/laws-of-reflection](https://go.dev/blog/laws-of-reflection)

Interfaces are satisfied **implicitly** (no `implements` keyword).  
Source: [https://go.dev/doc/faq#implements_declaration](https://go.dev/doc/faq#implements_declaration)

---

### 9. `:=` shadowing

`:=` declares new variables in the current block. It may **redeclare** variables already declared in the **same** block only in a multi-variable `:=` where at least one name is new (redeclaration assigns, does not create a second binding).

Source: [https://go.dev/ref/spec#Short_variable_declarations](https://go.dev/ref/spec#Short_variable_declarations)

**Shadowing gotcha:** `:=` in an inner block (`if`, `for`, nested function) creates a **new** variable that hides the outer one. The outer binding is unchanged. Spec example: a named result `err` shadowed by `err := ...` can make a bare `return` invalid / surprising.

Source: [https://go.dev/ref/spec#Return_statements](https://go.dev/ref/spec#Return_statements) (shadowed named result)

Unlike JS `let` TDZ quirks or PHP/Python assignment in nested scopes, Go’s block rules are strict—and `:=` makes accidental inner declarations easy.

---

### 10. `defer` semantics (LIFO, arg evaluation timing, loops)

Rules (blog + spec):

1. **Arguments are evaluated when `defer` executes**, not when the deferred call runs.  
2. Deferred calls run **LIFO** immediately before the surrounding function returns.  
3. Deferred functions may read/assign **named result parameters**.

Sources: [https://go.dev/blog/defer-panic-and-recover](https://go.dev/blog/defer-panic-and-recover), [https://go.dev/ref/spec#Defer_statements](https://go.dev/ref/spec#Defer_statements)

**Loop gotcha:** `defer` is function-scoped, not block-scoped. Deferring inside a loop schedules one call per iteration, all running at function exit (resource exhaustion / surprising ordering). Prefer closing inside the loop body, or factor work into a function so each iteration has its own defers.

Source: [https://go.dev/doc/effective_go#defer](https://go.dev/doc/effective_go#defer) (function-based, not block-based)

---

### 11. No ternary / default args / overloading

| Missing feature | Official stance |
| --- | --- |
| Ternary `?:` | Intentionally omitted; use `if`/`else` |
| Method / operator overloading | Not supported (simplicity) |
| Default parameter values | Not part of the language (no such parameter syntax in the spec) |

Sources: [https://go.dev/doc/faq#Does_Go_have_a_ternary_form](https://go.dev/doc/faq#Does_Go_have_a_ternary_form) (FAQ: “Why does Go not have the ?: operator?”), [https://go.dev/doc/faq#overloading](https://go.dev/doc/faq#overloading), language grammar: [https://go.dev/ref/spec](https://go.dev/ref/spec)

Workarounds used in stdlib style: multiple functions (`context.WithTimeout` vs `WithDeadline`), variadic options, or zero-value + explicit fields—not default args.

---

### 12. Switch (no auto fallthrough)

In an expression `switch`, each `case` ends that switch unless the last statement is an explicit **`fallthrough`**. Cases do **not** fall through by default (unlike C). `fallthrough` is **not** allowed in type switches.

Source: [https://go.dev/ref/spec#Switch_statements](https://go.dev/ref/spec#Switch_statements)

---

### 13. Concurrency: goroutines, channels, races, context, loop var capture (pre-1.22 vs 1.22+)

**Model:** Prefer “Do not communicate by sharing memory; instead, share memory by communicating” (channels passing ownership / references).

Sources: [https://go.dev/blog/share-memory-by-communicating](https://go.dev/blog/share-memory-by-communicating), [https://go.dev/doc/codewalk/sharemem](https://go.dev/doc/codewalk/sharemem), [https://go.dev/doc/effective_go#concurrency](https://go.dev/doc/effective_go#concurrency)

**Data races:** A race is concurrent access to the same location where at least one access is a write (unless `sync/atomic`). Use `go test -race` / `go build -race`. Maps are a common race site.

Sources: [https://go.dev/ref/mem](https://go.dev/ref/mem), [https://go.dev/doc/articles/race_detector](https://go.dev/doc/articles/race_detector)

**Context:** Carries deadlines, cancellation, and request-scoped values across API boundaries. Propagate `ctx`; call the `CancelFunc` from `WithCancel` / `WithTimeout` (failing to cancel can leak until the parent cancels). Safe for concurrent use.

Sources: [https://go.dev/blog/context](https://go.dev/blog/context), [https://pkg.go.dev/context](https://pkg.go.dev/context)

**Loop variable capture — version-specific (important correction vs older folklore):**

| Language version | Semantics |
| --- | --- |
| **Before Go 1.22** (or modules with `go` &lt; 1.22) | `for` loop variables were **per-loop**, reused each iteration → closures/goroutines often saw the **last** value |
| **Go 1.22+** (modules declaring `go 1.22` or later) | Each iteration creates **new** variables → accidental sharing fixed for typical closure/goroutine patterns |

The change is gated by the module’s `go` line (not merely the toolchain binary). Go 1.21 offered a `GOEXPERIMENT=loopvar` preview.

Sources:

- [https://go.dev/blog/loopvar-preview](https://go.dev/blog/loopvar-preview)  
- [https://go.dev/doc/go1.22](https://go.dev/doc/go1.22)  
- [https://go.dev/blog/go1.22](https://go.dev/blog/go1.22)  
- [https://go.dev/wiki/LoopvarExperiment](https://go.dev/wiki/LoopvarExperiment)  
- [https://go.dev/wiki/CommonMistakes](https://go.dev/wiki/CommonMistakes) (notes pre-1.22)  
- FAQ: [https://go.dev/doc/faq#closures_and_goroutines](https://go.dev/doc/faq#closures_and_goroutines)

---

### 14. Packages, circular imports, `init()`

- Import graph is **acyclic by construction**: imported packages initialize before the importer; cycles are not allowed (no cyclic initialization dependencies).  
- Package-level vars initialize by dependency order; then all `init` functions run in source order as presented to the compiler.  
- `init` has no parameters/results, cannot be referenced, may appear multiple times per package.  
- Unused imports are a **compile error**.

Sources: [https://go.dev/ref/spec#Program_initialization_and_execution](https://go.dev/ref/spec#Program_initialization_and_execution), [https://go.dev/doc/faq#unused_variables_and_imports](https://go.dev/doc/faq#unused_variables_and_imports)

Side-effect-only import: `import _ "pkg"` (e.g. `net/http/pprof` registration).  
Source: [https://go.dev/doc/effective_go#blank_import](https://go.dev/doc/effective_go#blank_import)

---

### 15. JSON tags / exported fields / `omitempty` zero-value behavior

- Only **exported** fields are marshaled/unmarshaled.  
- Field names / tags control JSON keys; unexported fields are ignored.  
- **`omitempty`**: omit if “empty”: `false`, `0`, nil pointer, nil interface, or array/slice/map/string of **length zero**.  
- **`omitzero`** (newer option in `encoding/json`): omit if the field is a **zero value** for its type (or if it has `IsZero() bool`). If both options are set, omit if either applies.

Sources: [https://go.dev/blog/json](https://go.dev/blog/json), package docs in [https://go.dev/src/encoding/json/encode.go](https://go.dev/src/encoding/json/encode.go) (struct tag comment block), [https://pkg.go.dev/encoding/json](https://pkg.go.dev/encoding/json)

**Gotcha:** `omitempty`’s “empty” list is **not** identical to “zero value of every type” (e.g. historical surprise around empty structs / non-length-based zeros). Prefer reading current `encoding/json` docs; use `omitzero` when you mean “zero value”.

---

### 16. Comparability (structs vs slices/maps)

Comparable with `==` / `!=`: among others, bools, numbers, strings, pointers, channels, interfaces (with panics if dynamic type is not comparable), structs **if all fields are comparable**, arrays **if element type is comparable**.

**Not comparable:** slices, maps, functions (except comparison to `nil`).

Therefore slices/maps **cannot be map keys**.

Sources: [https://go.dev/ref/spec#Comparison_operators](https://go.dev/ref/spec#Comparison_operators), [https://go.dev/blog/go-maps-in-action](https://go.dev/blog/go-maps-in-action)

---

### 17. What replaces classes / inheritance / exceptions / optional chaining

| From JS/TS, PHP, Python | In Go |
| --- | --- |
| Classes + inheritance | Structs + methods; **composition**; interfaces satisfied implicitly — **no type hierarchy** |
| `implements` / ABC declarations | Implicit satisfaction of interfaces |
| Exceptions / try-catch for control flow | `error` values; rare `panic`/`recover` |
| `obj?.prop` optional chaining | Explicit `if p != nil`, comma-ok maps, or helpers — **no** `?.` operator in the language |
| Default args / overloads | Separate functions, variadic params, config structs |
| Monkey-patch / open classes | Not applicable; interfaces + wrapping |

Sources: [https://go.dev/doc/faq#inheritance](https://go.dev/doc/faq#inheritance), [https://go.dev/doc/faq#implements_declaration](https://go.dev/doc/faq#implements_declaration), [https://go.dev/doc/faq#exceptions](https://go.dev/doc/faq#exceptions), [https://go.dev/doc/faq#overloading](https://go.dev/doc/faq#overloading), keywords/operators in [https://go.dev/ref/spec](https://go.dev/ref/spec)

---

## Part B — Go memory management model

### 18. GC overview (concurrent tracing mark-and-sweep; not reference counting)

Go uses **tracing** garbage collection (follow pointers to find live objects), specifically a **mark-sweep** collector. It is **not** reference counting. On multiprocessors the collector runs concurrently with the program (implementation evolves; FAQ still describes mark-and-sweep with parallel collection).

Sources:

- [https://go.dev/doc/gc-guide](https://go.dev/doc/gc-guide) (tracing vs refcounting; mark-sweep phases)  
- [https://go.dev/doc/faq#garbage_collection](https://go.dev/doc/faq#garbage_collection)

---

### 19. Stack vs heap and escape analysis

- Locals whose lifetime the compiler can bound often live on the **goroutine stack** (not GC-managed in the usual sense).  
- Values that must outlive the frame (or whose lifetime cannot be proven) **escape to the heap** and are GC-managed.  
- Taking an address is a candidate for heap allocation; escape analysis may still keep some address-taken vars on the stack.  
- Escape decisions change across compiler versions — do not memorize brittle rules; measure.

Sources: [https://go.dev/doc/faq#stack_or_heap](https://go.dev/doc/faq#stack_or_heap), [https://go.dev/doc/gc-guide](https://go.dev/doc/gc-guide)

---

### 20. Values, pointers, copies; slice / map / chan / interface / string headers

From the spec’s representation rules:

| Kind | What copying the variable copies |
| --- | --- |
| Basic types, arrays, structs | The full value (deep for nested value fields) |
| Pointers | The pointer (shared pointee) |
| Slices | Header (ptr/len/cap) — **shared array** |
| Maps / channels | Reference to shared structure |
| Interfaces | Interface pair; may share or contain data depending on dynamic type |
| Strings | String header (data pointer + len); data treated as immutable |

Sources: [https://go.dev/ref/spec#Representation_of_values](https://go.dev/ref/spec#Representation_of_values), [https://go.dev/doc/faq#pass_by_value](https://go.dev/doc/faq#pass_by_value), [https://go.dev/blog/strings](https://go.dev/blog/strings)

---

### 21. Slice aliasing and retaining large backing arrays

Re-slicing shares the array. A tiny leftover slice can pin a huge buffer in memory until all references are gone. Fix: `copy` into a new slice sized to the keep-set.

Source: [https://go.dev/blog/slices-intro](https://go.dev/blog/slices-intro)

Same class of issue as keeping a substring that shares storage — but Go **strings** are immutable; the classic leak pattern is especially about **`[]byte` / slices**.

---

### 22. Allocation pressure; tools (`-gcflags=-m`, pprof, `AllocsPerRun`)

| Tool | Use |
| --- | --- |
| `go build -gcflags=-m=3 <pkg>` | Compiler escape / optimization decisions |
| `runtime/pprof` / `net/http/pprof` heap profile | Where allocations come from |
| `testing.AllocsPerRun` | Average mallocs per call in tests (sets `GOMAXPROCS` to 1 during measurement) |
| `runtime.ReadMemStats` / `debug.ReadGCStats` | High-level heap / GC stats |
| `runtime.NumGoroutine` | Spot goroutine growth / leaks |

Sources: [https://go.dev/doc/gc-guide](https://go.dev/doc/gc-guide), [https://go.dev/doc/diagnostics](https://go.dev/doc/diagnostics), [https://go.dev/src/testing/allocs.go](https://go.dev/src/testing/allocs.go)

---

### 23. `GOGC` and `GOMEMLIMIT`

| Knob | Role | Since / notes |
| --- | --- | --- |
| **`GOGC`** | Trade-off between GC CPU and heap overhead. Target heap grows roughly with live heap × `GOGC/100` (formula includes GC roots as of Go **1.18**). Default conceptual midpoint is 100. `GOGC=off` / `SetGCPercent(-1)` disables GC **unless** a memory limit applies. | Long-standing; root-set change in 1.18 |
| **`GOMEMLIMIT`** | Soft runtime memory limit so GC can run more often under pressure instead of OOM on spikes. Set via env or `runtime/debug.SetMemoryLimit`. | Added in Go **1.19** |

Doubling `GOGC` roughly doubles heap overhead and halves GC CPU cost (guide’s rule of thumb).

Source: [https://go.dev/doc/gc-guide](https://go.dev/doc/gc-guide)

**Correction vs older advice:** “only tune GOGC” is incomplete on modern Go — soft memory limits are first-class.

---

### 24. Goroutine stacks; goroutine leaks as memory leaks

New goroutines start with small, **growable** stacks (a few kilobytes initially; runtime grows/shrinks). Hundreds of thousands of goroutines can be practical because stacks are cheap compared to OS threads.

Source: [https://go.dev/doc/faq#goroutines](https://go.dev/doc/faq#goroutines)

A **blocked goroutine** that never becomes runnable retains its stack and any heap objects it references. That is a common “memory leak” pattern (also: forgetting `cancel` on derived contexts). Diagnostics: goroutine profiles, `NumGoroutine`, execution tracer.

Sources: [https://go.dev/doc/diagnostics](https://go.dev/doc/diagnostics), [https://pkg.go.dev/context](https://pkg.go.dev/context) (CancelFunc leak note)

---

### 25. Formal Go memory model (synchronization / data races) — distinct from GC

The **Go memory model** defines when one event **happens before** another and what programmers may assume about concurrent reads/writes. It is about **visibility and synchronization**, not about when the GC reclaims memory.

- Data-race-free programs behave as if sequentially consistent (DRF-SC).  
- Channels, mutexes, `sync/atomic`, etc. establish happens-before edges.  
- GC tracing is a separate runtime subsystem ([gc-guide](https://go.dev/doc/gc-guide)).

Source: [https://go.dev/ref/mem](https://go.dev/ref/mem)

---

### 26. `string` ↔ `[]byte` conversions

- Spec: converting `string` → `[]byte` yields a slice whose elements are the string’s bytes; converting `[]byte` → `string` yields a string with those bytes.  
- Strings are immutable / read-only; byte slices are mutable. Preserving those semantics generally requires **distinct backing storage** (a copy) for the conversion result, though implementations may optimize when they can prove safety.

Sources: [https://go.dev/ref/spec#Conversions_to_and_from_a_string_type](https://go.dev/ref/spec#Conversions_to_and_from_a_string_type), [https://go.dev/blog/strings](https://go.dev/blog/strings), [https://go.dev/doc/faq#pass_by_value](https://go.dev/doc/faq#pass_by_value) (implementations may optimize without changing semantics)

For hot paths, measure with `AllocsPerRun` / escape analysis rather than assuming every conversion allocates in every Go version.

---

### 27. Finalizers (`SetFinalizer`) — nondeterministic; prefer `defer Close`

`runtime.SetFinalizer` runs a function after the GC finds an object unreachable — **at an arbitrary later time**, on a single finalizer goroutine. **Not guaranteed** before program exit; not guaranteed for zero-size objects or some linker-allocated package-level objects; cycles with finalizers may never be collected.

Docs recommend **deterministic cleanup** (`Close` + `defer`) for non-memory resources. Finalizers (and newer `AddCleanup`) are best-effort / last resort. GC guide: prefer cleanups to finalizers where available; prefer explicit `Close`.

Sources: [https://pkg.go.dev/runtime#SetFinalizer](https://pkg.go.dev/runtime#SetFinalizer), [https://go.dev/src/runtime/mfinal.go](https://go.dev/src/runtime/mfinal.go), [https://go.dev/doc/gc-guide](https://go.dev/doc/gc-guide), [https://go.dev/blog/defer-panic-and-recover](https://go.dev/blog/defer-panic-and-recover)

---

### 28. Comparison table — JS / Python / PHP habits → Go reality

| Habit from JS / Python / PHP | Go reality | Primary source |
| --- | --- | --- |
| Throw / raise for ordinary errors | Return `error`; check explicitly | [error-handling blog](https://go.dev/blog/error-handling-and-go) |
| `null` / `None` / `undefined` everywhere | Typed zeros; `nil` only for reference-ish types | [spec zero value](https://go.dev/ref/spec#The_zero_value) |
| Objects are references | Structs copy; use pointers or slice/map headers | [FAQ pass by value](https://go.dev/doc/faq#pass_by_value) |
| Classes + inheritance | Structs + interfaces + composition | [FAQ inheritance](https://go.dev/doc/faq#inheritance) |
| `public` / `private` keywords | Capitalization export | [spec exported identifiers](https://go.dev/ref/spec#Exported_identifiers) |
| Growable arrays / lists share freely | Slice headers share arrays; watch aliasing & retention | [slices blog](https://go.dev/blog/slices-intro) |
| Dict/object missing key quirks | Missing map key → **zero value**; write to nil map **panics** | [maps in action](https://go.dev/blog/go-maps-in-action) |
| Strings as characters | Bytes + runes; `len` is bytes | [strings blog](https://go.dev/blog/strings) |
| GC is “free” / refcounting mental model | Tracing mark-sweep; escape analysis matters | [gc-guide](https://go.dev/doc/gc-guide) |
| Threads + shared mutable state first | Goroutines + channels; still need sync for shared memory | [share memory blog](https://go.dev/blog/share-memory-by-communicating) |
| `finally` / context managers | `defer` (function scope, LIFO) | [defer blog](https://go.dev/blog/defer-panic-and-recover) |
| Optional chaining / ternaries | Explicit `if`; no `?:` | [FAQ ternary](https://go.dev/doc/faq#Does_Go_have_a_ternary_form) |

---

### 29. Practical rules of thumb (grounded in official docs)

1. **Return and check `error`s**; reserve `panic` for true anomalies. ([error blog](https://go.dev/blog/error-handling-and-go), [effective_go](https://go.dev/doc/effective_go#errors))  
2. **Return `nil` of type `error`**, not a nil concrete pointer in an interface. ([FAQ nil error](https://go.dev/doc/faq#nil_error))  
3. **`make` maps before write**; treat concurrent map use as a race. ([maps blog](https://go.dev/blog/go-maps-in-action), [race detector](https://go.dev/doc/articles/race_detector))  
4. **Assign `append`’s result**; don’t assume capacity never reallocates. ([slices blog](https://go.dev/blog/slices-intro))  
5. **Copy out** small slices you retain from large buffers. ([slices blog](https://go.dev/blog/slices-intro))  
6. **`defer Close` / unlock** next to acquisition; don’t rely on finalizers. ([defer blog](https://go.dev/blog/defer-panic-and-recover), [SetFinalizer](https://pkg.go.dev/runtime#SetFinalizer))  
7. **Propagate `context.Context`** and call cancel functions. ([context blog](https://go.dev/blog/context))  
8. Know your module’s **`go` version** for loop-variable semantics (≥ 1.22). ([go1.22 notes](https://go.dev/doc/go1.22))  
9. For memory: design for useful **zero values**, reduce unnecessary heap escape, profile before tuning, then consider **`GOGC` / `GOMEMLIMIT`**. ([gc-guide](https://go.dev/doc/gc-guide), [FAQ stack/heap](https://go.dev/doc/faq#stack_or_heap))  
10. Treat **data races** as bugs even if “it works on my machine”; use `-race`. ([memory model](https://go.dev/ref/mem), [race detector](https://go.dev/doc/articles/race_detector))

---

## Uncertainties / limits of this research

- **Conversation claims:** No prior chat transcript was available in-repo; this note follows the requested topic list and corrects **common** outdated statements against current primary docs (especially **Go 1.22 loop scoping**, **GOMEMLIMIT since 1.19**, **GOGC root-set accounting since 1.18**, and **`omitzero` vs `omitempty`**).  
- **`encoding/json` tag semantics** are taken from the stdlib source comment in `encode.go` / pkg.go.dev as of research time; patch releases can refine edge cases — re-read [https://pkg.go.dev/encoding/json](https://pkg.go.dev/encoding/json) when upgrading Go.  
- **`string` ↔ `[]byte` physical copying:** the spec defines conversion results; immutability implies separate mutable storage in the general case, but the compiler/runtime may elide copies when semantics allow. Treat allocation as likely until measured.  
- **GC implementation details** (pacer, write barriers, exact pause behavior) evolve; prefer [https://go.dev/doc/gc-guide](https://go.dev/doc/gc-guide) over older FAQ one-liners for tuning.  
- **`AddCleanup` vs `SetFinalizer`:** current runtime docs steer new code toward cleanups; availability depends on the Go version this repo pins — verify against your toolchain.  
- **pkg.go.dev** intermittently returned HTTP 403 during fetches; claims were cross-checked via `go.dev/src/...` and successfully retrieved package pages where possible.
