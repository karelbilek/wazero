package experimental_test

import (
	"bufio"
	"context"
	_ "embed"
	"strings"
	"testing"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/experimental"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/internal/platform"
	"github.com/tetratelabs/wazero/internal/testing/dwarftestdata"
	"github.com/tetratelabs/wazero/internal/testing/require"
)

func TestWithDWARFBasedStackTrace(t *testing.T) {
	ctx := context.Background()
	require.False(t, experimental.DWARFBasedStackTraceEnabled(ctx))
	ctx = experimental.WithDWARFBasedStackTrace(ctx)
	require.True(t, experimental.DWARFBasedStackTraceEnabled(ctx))

	type testCase struct {
		name string
		r    wazero.Runtime
	}

	tests := []testCase{{name: "interpreter", r: wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfigInterpreter())}}

	if platform.CompilerSupported() {
		tests = append(tests, testCase{
			name: "compiler", r: wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfigCompiler()),
		})
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			r := tc.r
			defer r.Close(ctx) // This closes everything this Runtime created.
			wasi_snapshot_preview1.MustInstantiate(ctx, r)

			for _, lang := range []struct {
				name string
				bin  []byte
				exp  string
			}{
				{
					name: "tinygo",
					bin:  dwarftestdata.TinyGoWasm,
					exp: `module[] function[_start] failed: wasm error: unreachable
wasm stack trace:
	.runtime._panic(i32)
		0x16e2: /runtime_tinygowasm.go:73:6 (inlined)
		        /panic.go:52:7
	.c()
		0x1911: /main.go:19:7
	.b()
		0x1901: /main.go:14:3
	.a()
		0x18f7: /main.go:9:3
	.main.main()
		0x18ed: /main.go:4:3
	.runtime.run()
		0x18cc: /scheduler_none.go:26:10
	._start()
		0x18b6: /runtime_wasm_wasi.go:22:5`,
				},
				{
					name: "zig",
					bin:  dwarftestdata.ZigWasm,
					exp: `module[] function[_start] failed: wasm error: unreachable
wasm stack trace:
	.os.abort()
		0x1b3: /os.zig:552:9
	.builtin.default_panic(i32,i32,i32,i32)
		0x86: /builtin.zig:787:25
	.main.main() i32
		0x25: main.zig:10:5 (inlined)
		      main.zig:6:5 (inlined)
		      main.zig:2:5
	._start()
		0x1c6: /start.zig:614:37 (inlined)
		       /start.zig:240:42`,
				},
				{
					name: "rust",
					bin:  dwarftestdata.RustWasm,
					exp: `module[] function[_start] failed: wasm error: unreachable
wasm stack trace:
	.__rust_start_panic(i32) i32
		0xc474: /index.rs:286:39 (inlined)
		        /const_ptr.rs:870:18 (inlined)
		        /index.rs:286:39 (inlined)
		        /mod.rs:1630:46 (inlined)
		        /mod.rs:405:20 (inlined)
		        /mod.rs:1630:46 (inlined)
		        /mod.rs:1548:18 (inlined)
		        /iter.rs:1478:30 (inlined)
		        /count.rs:74:18
	.rust_panic(i32,i32)
		0xa3f8: /validations.rs:57:19 (inlined)
		        /validations.rs:57:19 (inlined)
		        /iter.rs:140:15 (inlined)
		        /iter.rs:140:15 (inlined)
		        /iterator.rs:330:13 (inlined)
		        /iterator.rs:377:9 (inlined)
		        /mod.rs:1455:35
	.std::panicking::rust_panic_with_hook::h93e119628869d575(i32,i32,i32,i32,i32)
		0x42df: /alloc.rs:244:22 (inlined)
		        /alloc.rs:244:22 (inlined)
		        /alloc.rs:342:9 (inlined)
		        /mod.rs:487:1 (inlined)
		        /mod.rs:487:1 (inlined)
		        /mod.rs:487:1 (inlined)
		        /mod.rs:487:1 (inlined)
		        /mod.rs:487:1 (inlined)
		        /panicking.rs:292:17 (inlined)
		        /panicking.rs:292:17
	.std::panicking::begin_panic_handler::{{closure}}::h2b8c0798e533b227(i32,i32,i32)
		0xaa8c: /mod.rs:362:12 (inlined)
		        /mod.rs:1257:22 (inlined)
		        /mod.rs:1235:21 (inlined)
		        /mod.rs:1214:26
	.std::sys_common::backtrace::__rust_end_short_backtrace::h030a533bc034da65(i32)
		0xc144: /mod.rs:188:26
	.rust_begin_unwind(i32)
		0xb7df: /mod.rs:1629:9 (inlined)
		        /builders.rs:199:17 (inlined)
		        /result.rs:1352:22 (inlined)
		        /builders.rs:187:23
	.core::panicking::panic_fmt::hb1bfc4175f838eff(i32,i32)
		0xbd3d: /mod.rs:1384:17
	.main::main::hfd44f54575e6bfdf()
		0xad2c: /memchr.rs
	.core::ops::function::FnOnce::call_once::h87e5f77996df3e28(i32)
		0xbd61: /mod.rs
	.std::sys_common::backtrace::__rust_begin_short_backtrace::h7ca17eb6aa97f768(i32)
		0xbd95: /mod.rs:1504:35 (inlined)
		        /mod.rs:1407:36
	.std::rt::lang_start::{{closure}}::he4aa401e76315dfe(i32) i32
		0xae9a: /location.rs:196:6
	.std::rt::lang_start_internal::h3c39e5d3c278a90f(i32,i32,i32,i32) i32
	.std::rt::lang_start::h779801844bd22a3c(i32,i32,i32) i32
		0xab94: /mod.rs:1226:2
	.__original_main() i32
		0xc0ae: /methods.rs:1677:13 (inlined)
		        /mod.rs:165:24 (inlined)
		        /mod.rs:165:24
	._start()
		0xc10f: /mod.rs:187
	._start.command_export()
		0xc3de: /iterator.rs:2414:21 (inlined)
		        /map.rs:124:9 (inlined)
		        /accum.rs:42:17 (inlined)
		        /iterator.rs:3347:9 (inlined)
		        /count.rs:135:5 (inlined)
		        /count.rs:135:5 (inlined)
		        /count.rs:71:21`,
				},
			} {
				t.Run(lang.name, func(t *testing.T) {
					if len(lang.bin) == 0 {
						t.Skip()
					}
					compiled, err := r.CompileModule(ctx, lang.bin)
					require.NoError(t, err)

					// Use context.Background to ensure that DWARF is a compile-time option.
					_, err = r.InstantiateModule(context.Background(), compiled, wazero.NewModuleConfig())
					require.Error(t, err)

					errStr := err.Error()

					// Since stack traces change where the binary is compiled, we sanitize each line
					// so that it doesn't contain any file system dependent info.
					scanner := bufio.NewScanner(strings.NewReader(errStr))
					scanner.Split(bufio.ScanLines)
					var sanitizedLines []string
					for scanner.Scan() {
						line := scanner.Text()
						start, last := strings.Index(line, "/"), strings.LastIndex(line, "/")
						if start >= 0 {
							l := len(line) - last
							buf := []byte(line)
							copy(buf[start:], buf[last:])
							line = string(buf[:start+l])
						}
						sanitizedLines = append(sanitizedLines, line)
					}

					sanitizedTraces := strings.Join(sanitizedLines, "\n")
					require.Equal(t, sanitizedTraces, lang.exp)
				})
			}
		})
	}
}
