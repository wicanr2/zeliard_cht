/* SPDX-License-Identifier: LGPL-2.1-or-later
 * Minimal allocation ABI for Nuked-OPL3, 2026 Zeliard contributors. */
#include "opl3.h"
#include <stdlib.h>
#if defined(_WIN32)
#define API __declspec(dllexport)
#else
#define API __attribute__((visibility("default")))
#endif
API void *zopl_new(uint32_t rate) { opl3_chip *c = calloc(1, sizeof(*c)); if(c) OPL3_Reset(c, rate); return c; }
API void zopl_free(void *c) { free(c); }
API void zopl_write(void *c, uint16_t reg, uint8_t v) { OPL3_WriteRegBuffered(c, reg, v); }
API double zopl_sample(void *c) { int16_t b[2]; OPL3_GenerateResampled(c, b); return ((double)b[0] + b[1]) / 65536.0; }
