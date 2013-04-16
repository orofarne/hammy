#pragma once

#ifdef __cplusplus
extern "C" {
#endif

#if GLIB_MAJOR_VERSION != 2
#error "invalig glib version"
#endif

#if GLIB_MINOR_VERSION < 34
/*
 * Backport of G_DEFINE_QUARK macro from glib 2.34
 */
#define G_DEFINE_QUARK(QN, q_n)                                         \
GQuark                                                                  \
q_n##_quark (void)                                                      \
{                                                                       \
  static GQuark q;                                                      \
                                                                        \
  if G_UNLIKELY (q == 0)                                                \
    q = g_quark_from_static_string (#QN);                               \
                                                                        \
  return q;                                                             \
}
#endif


#ifdef __cplusplus
} // extern "C"
#endif
