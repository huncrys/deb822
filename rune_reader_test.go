package deb822_test

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"oaklab.hu/debian/deb822"
)

func TestRuneReader_PeekRune(t *testing.T) {
	input := `A string with emojis: 😀, 🚀, 🍕`
	reader := deb822.NewRuneReader(strings.NewReader(input))
	prn, psize, perr := reader.PeekRune()
	require.NoError(t, perr)
	require.Equal(t, 'A', prn)
	require.Equal(t, 1, psize)

	rn, size, err := reader.ReadRune()
	require.NoError(t, err)
	require.Equal(t, prn, rn)
	require.Equal(t, psize, size)

	_, err = io.ReadAll(reader)
	require.NoError(t, err)

	_, _, err = reader.PeekRune()
	require.Error(t, err)
}

func TestRuneReader_Discard(t *testing.T) {
	input := "   \t\n  Hello, World!"
	reader := deb822.NewRuneReader(strings.NewReader(input))

	reader.DiscardSpace()

	rn, _, err := reader.ReadRune()
	require.NoError(t, err)
	require.Equal(t, 'H', rn)

	reader.DiscardRune()
	rn, _, err = reader.ReadRune()
	require.NoError(t, err)
	require.Equal(t, 'l', rn)
}
