/*
 * glibc symbol-version compatibility shim (Linux x86-64 release builds).
 *
 * Problem: glibc 2.38 introduced __isoc23_* variants of strtol/scanf and the
 * headers redirect calls to them, so anything compiled on a glibc >= 2.38 host
 * requires GLIBC_2.38 at runtime even when the code is ancient C. That silently
 * raises the minimum distro of our release binaries every time the CI runner
 * image is upgraded.
 *
 * Fix: bind those calls back to the original symbols, which have existed since
 * GLIBC_2.2.5 and behave identically for every input Ikemen and FFmpeg use. The
 * only difference in C23 is that strtol accepts a "0b" binary prefix; nothing
 * here relies on that.
 *
 * Deliberately narrow. A blanket header such as wheybags/glibc_version_header
 * cannot be used here: it also emits .symver lines for removed interfaces like
 * sysctl, which makes FFmpeg's configure link test succeed and set
 * HAVE_SYSCTL=1, after which libavutil/cpu.c includes <sys/sysctl.h> - a header
 * glibc deleted in 2.32 - and the build fails. Only add symbols below that are
 * genuinely version-bumped aliases of a still-present function.
 *
 * Applied via -include from build/build.sh for GOOS=linux.
 *
 * Two constraints this file must respect, both learned from CI failures:
 *
 *  1. It is force-included into EVERY translation unit, including .S assembly
 *     files (runtime/cgo has some). The assembler chokes on C syntax, so
 *     everything is guarded on !defined(__ASSEMBLER__).
 *
 *  2. It must NOT include <features.h>. Being force-included, it runs before
 *     the TU's own "#define _GNU_SOURCE", and features.h latches the feature
 *     macros on first inclusion - which hides GNU extensions such as
 *     setresgid/setresuid/pthread_getattr_np and breaks runtime/cgo. So the
 *     glibc version cannot be tested here; the .symver directives are emitted
 *     unconditionally and build.sh decides when to apply the header.
 *
 * Emitting .symver for a symbol the libc lacks is harmless: the assembler only
 * records an alias, and nothing references it unless the compiler generated a
 * call to that name in the first place.
 */

#ifndef IKEMEN_GLIBC_COMPAT_H
#define IKEMEN_GLIBC_COMPAT_H

#if defined(__linux__) && defined(__x86_64__) && !defined(__ASSEMBLER__)

/* strtol family: __isoc23_* -> GLIBC_2.2.5 originals */
__asm__(".symver __isoc23_strtol,strtol@GLIBC_2.2.5");
__asm__(".symver __isoc23_strtoll,strtoll@GLIBC_2.2.5");
__asm__(".symver __isoc23_strtoul,strtoul@GLIBC_2.2.5");
__asm__(".symver __isoc23_strtoull,strtoull@GLIBC_2.2.5");
__asm__(".symver __isoc23_strtoimax,strtoimax@GLIBC_2.2.5");
__asm__(".symver __isoc23_strtoumax,strtoumax@GLIBC_2.2.5");

/* wide-char strtol family */
__asm__(".symver __isoc23_wcstol,wcstol@GLIBC_2.2.5");
__asm__(".symver __isoc23_wcstoll,wcstoll@GLIBC_2.2.5");
__asm__(".symver __isoc23_wcstoul,wcstoul@GLIBC_2.2.5");
__asm__(".symver __isoc23_wcstoull,wcstoull@GLIBC_2.2.5");
__asm__(".symver __isoc23_wcstoimax,wcstoimax@GLIBC_2.2.5");
__asm__(".symver __isoc23_wcstoumax,wcstoumax@GLIBC_2.2.5");

/* scanf family: C23 variants -> the C99 ones, which predate 2.38 */
__asm__(".symver __isoc23_scanf,__isoc99_scanf@GLIBC_2.7");
__asm__(".symver __isoc23_fscanf,__isoc99_fscanf@GLIBC_2.7");
__asm__(".symver __isoc23_sscanf,__isoc99_sscanf@GLIBC_2.7");
__asm__(".symver __isoc23_vscanf,__isoc99_vscanf@GLIBC_2.7");
__asm__(".symver __isoc23_vfscanf,__isoc99_vfscanf@GLIBC_2.7");
__asm__(".symver __isoc23_vsscanf,__isoc99_vsscanf@GLIBC_2.7");
__asm__(".symver __isoc23_wscanf,__isoc99_wscanf@GLIBC_2.7");
__asm__(".symver __isoc23_fwscanf,__isoc99_fwscanf@GLIBC_2.7");
__asm__(".symver __isoc23_swscanf,__isoc99_swscanf@GLIBC_2.7");
__asm__(".symver __isoc23_vwscanf,__isoc99_vwscanf@GLIBC_2.7");
__asm__(".symver __isoc23_vfwscanf,__isoc99_vfwscanf@GLIBC_2.7");
__asm__(".symver __isoc23_vswscanf,__isoc99_vswscanf@GLIBC_2.7");

/* glibc 2.38 added a second version of fmod/fmodf; use the original. */
__asm__(".symver fmod,fmod@GLIBC_2.2.5");
__asm__(".symver fmodf,fmodf@GLIBC_2.2.5");

#endif /* linux && x86_64 && !assembler */

#endif /* IKEMEN_GLIBC_COMPAT_H */
