#include "msgbuf.h"

#include "glib_extra.h"
#include <errno.h>
#include <string.h>

G_DEFINE_QUARK (hammy-msgbuf-error, hammy_msgbuf_error)

// see http://wiki.msgpack.org/display/MSGPACK/Format+specification

static void
hammy_msg_buf_invalid_type(GError **error, guchar type)
{
	GError *lerr = NULL;
	lerr = g_error_new (hammy_msgbuf_error_quark(), EBADMSG, "invalid type: %x", (int)type);
	g_propagate_error (error, lerr);
}

static gsize
hammy_msg_buf_read_elem (gpointer buf, goffset off, gsize size, GError **error);

static gsize
hammy_msg_buf_read_elems (gpointer buf, goffset off, gsize size, guint32 n, GError **error)
{
	guint32 i;
	gsize r, total = 0;

	for (i = 0; i < n; i++)
	{
		r = hammy_msg_buf_read_elem (buf, off, size, error);
		if (r == 0)
			return 0;
		off += r;
		total += r;
	}

	return total;
}

static gsize
hammy_msg_buf_read_elem (gpointer buf, goffset off, gsize size, GError **error)
{
	guchar type;
	guint32 n;
	gsize r;

	if (off >= size)
	{
		return 0;
	}

	type = *((guchar *)buf + off);
	off++;

	// Positive FixNum 	0xxxxxxx 	0x00 - 0x7f
	if (type <= 0x7f)
	{
		return 1;
	}
	// FixMap 	1000xxxx 	0x80 - 0x8f
	if (type >= 0x80 && type <= 0x8f)
	{
		n = type ^ 0x80;
		if (n == 0)
			return 1;
		r = hammy_msg_buf_read_elems(buf, off, size, n * 2, error);
		return r > 0 ? r + 1 : 0;
	}
	// FixArray 	1001xxxx 	0x90 - 0x9f
	if (type >= 0x90 && type <= 0x9f)
	{
		n = type ^ 0x90;
		if (n == 0)
			return 1;
		r = hammy_msg_buf_read_elems(buf, off, size, n, error);
		return r > 0 ? r + 1 : 0;
	}
	// FixRaw 	101xxxxx 	0xa0 - 0xbf
	if (type >= 0xa0 && type <= 0xbf)
	{
		n = type ^ 0xa0;
		return n + 1;
	}
	// nil 	11000000 	0xc0
	if (type == 0xc0)
	{
		return 1;
	}
	// false 	11000010 	0xc2
	if (type == 0xc2)
	{
		return 1;
	}
	// true 	11000011 	0xc3
	if (type == 0xc3)
	{
		return 1;
	}
	// float 	11001010 	0xca
	if (type == 0xca)
	{
		return 5;
	}
	// double 	11001011 	0xcb
	if (type == 0xcb)
	{
		return 9;
	}
	// uint 8 	11001100 	0xcc
	if (type == 0xcc)
	{
		return 2;
	}
	// uint 16 	11001101 	0xcd
	if (type == 0xcd)
	{
		return 3;
	}
	// uint 32 	11001110 	0xce
	if (type == 0xce)
	{
		return 5;
	}
	// uint 64 	11001111 	0xcf
	if (type == 0xcf)
	{
		return 9;
	}
	// int 8 	11010000 	0xd0
	if (type == 0xd0)
	{
		return 2;
	}
	// int 16 	11010001 	0xd1
	if (type == 0xd1)
	{
		return 3;
	}
	// int 32 	11010010 	0xd2
	if (type == 0xd2)
	{
		return 5;
	}
	// int 64 	11010011 	0xd3
	if (type == 0xd3)
	{
		return 9;
	}
	// raw 16 	11011010 	0xda
	if (type == 0xda)
	{
		if (off + 2 > size)
		{
			return 0;
		}
		n = GUINT16_FROM_BE (*(guint16 *)((char *)buf + off));
		return n + 3;
	}
	// raw 32 	11011011 	0xdb
	if (type == 0xdb)
	{
		if (off + 4 > size)
		{
			return 0;
		}
		n = GUINT32_FROM_BE (*(guint32 *)((char *)buf + off));
		return n + 5;
	}
	// array 16 	11011100 	0xdc
	if (type == 0xdc)
	{
		if (off + 2 > size)
		{
			return 0;
		}
		n = GUINT16_FROM_BE (*(guint16 *)((char *)buf + off));
		if (n == 0)
			return 3;
		r = hammy_msg_buf_read_elems(buf, off + 2, size, n, error);
		return r > 0 ? r + 3 : 0;
	}
	// array 32 	11011101 	0xdd
	if (type == 0xdd)
	{
		if (off + 4 > size)
		{
			return 0;
		}
		n = GUINT32_FROM_BE (*(guint32 *)((char *)buf + off));
		if (n == 0)
			return 5;
		r = hammy_msg_buf_read_elems(buf, off + 2, size, n, error);
		return r > 0 ? r + 5 : 0;
	}
	// map 16 	11011110 	0xde
	if (type == 0xde)
	{
		if (off + 2 > size)
		{
			return 0;
		}
		n = GUINT16_FROM_BE (*(guint16 *)((char *)buf + off));
		if (n == 0)
			return 3;
		r = hammy_msg_buf_read_elems(buf, off + 2, size, n * 2, error);
		return r > 0 ? r + 3 : 0;
	}
	// map 32 	11011111 	0xdf
	if (type == 0xdf)
	{
		if (off + 4 > size)
		{
			return 0;
		}
		n = GUINT32_FROM_BE (*(guint32 *)((char *)buf + off));
		if (n == 0)
			return 5;
		r = hammy_msg_buf_read_elems(buf, off + 2, size, n * 2, error);
		return r > 0 ? r + 5 : 0;
	}
	// Negative FixNum 	111xxxxx 	0xe0 - 0xff
	if (type >= 0xe0 && type <= 0xff)
	{
		return 1;
	}

	hammy_msg_buf_invalid_type(error, type);
	return 0;
}

gsize
hammy_msg_buf_read (gpointer buf, gsize size, GError **error)
{
	gsize r;

	r = hammy_msg_buf_read_elem(buf, 0, size, error);

	if (r > size)
		return 0;
	return r;
}