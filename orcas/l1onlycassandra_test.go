// Copyright 2017 Netflix, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package orcas_test

import (
	"bufio"
	"bytes"
	"fmt"
	"testing"

	"github.com/BarthV/memandra/orcas"
	"github.com/netflix/rend/common"
	"github.com/netflix/rend/protocol/textprot"
)

func TestL1OnlyCassandraOrca(t *testing.T) {
	// GET -> L1 HIT
	t.Run("Get", func(t *testing.T) {
		t.Run("L1Hit", func(t *testing.T) {
			h1 := &testHandler{
				responses: []common.GetResponse{
					{
						Key:  []byte("key"),
						Data: []byte("foo"),
					},
				},
			}
			h2 := &testHandler{}
			output := &bytes.Buffer{}

			l1only := orcas.L1OnlyCassandra(h1, h2, textprot.NewTextResponder(bufio.NewWriter(output)))

			err := l1only.Get(common.GetRequest{
				Keys:    [][]byte{[]byte("key")},
				Opaques: []uint32{0},
				Quiet:   []bool{false},
				NoopEnd: false,
			})
			if err != nil {
				t.Fatalf("Error should be nil, got %v", err)
			}

			out := string(output.Bytes())

			t.Logf(out)
			gold := "VALUE key 0 3\r\nfoo\r\nEND\r\n"

			if out != gold {
				t.Fatalf("Expected response '%v' but got '%v'", gold, out)
			}

			h1.verifyEmpty(t)
			h2.verifyEmpty(t)
		})

		// GET -> L1 MISS
		t.Run("L1Miss", func(t *testing.T) {
			h1 := &testHandler{
				errors: []error{common.ErrKeyNotFound},
			}
			h2 := &testHandler{}
			output := &bytes.Buffer{}

			l1only := orcas.L1OnlyCassandra(h1, h2, textprot.NewTextResponder(bufio.NewWriter(output)))

			err := l1only.Get(common.GetRequest{
				Keys:    [][]byte{[]byte("key")},
				Opaques: []uint32{0},
				Quiet:   []bool{false},
				NoopEnd: false,
			})
			if err != common.ErrKeyNotFound {
				t.Fatalf("Error should be %s, got %v", common.ErrKeyNotFound, err)
			}

			out := string(output.Bytes())

			t.Logf(out)
			gold := ""

			if out != gold {
				t.Fatalf("Expected response '%v' but got '%v'", gold, out)
			}

			h1.verifyEmpty(t)
			h2.verifyEmpty(t)
		})
	})

	t.Run("Set", func(t *testing.T) {
		// SET -> L1 OJ
		t.Run("L1SetSuccess", func(t *testing.T) {
			h1 := &testHandler{
				errors: []error{nil},
			}
			h2 := &testHandler{}
			output := &bytes.Buffer{}

			l1only := orcas.L1OnlyCassandra(h1, h2, textprot.NewTextResponder(bufio.NewWriter(output)))

			err := l1only.Set(common.SetRequest{})
			if err != nil {
				t.Fatalf("Error should be nil, got %v", err)
			}

			out := string(output.Bytes())

			t.Logf(out)

			if out != "STORED\r\n" {
				t.Fatalf("Expected response 'STORED\\r\\n' but got '%v'", out)
			}

			h1.verifyEmpty(t)
			h2.verifyEmpty(t)
		})

		// SET -> L1 FAIL
		t.Run("L1SetFailure", func(t *testing.T) {
			h1 := &testHandler{
				errors: []error{common.ErrInternal},
			}
			h2 := &testHandler{}
			output := &bytes.Buffer{}

			l1only := orcas.L1OnlyCassandra(h1, h2, textprot.NewTextResponder(bufio.NewWriter(output)))

			err := l1only.Set(common.SetRequest{})
			if err != common.ErrInternal {
				t.Fatalf("Error should be %s, got %v", common.ErrInternal, err)
			}

			out := string(output.Bytes())

			t.Logf(out)

			if out != "" {
				t.Fatalf("Expected response 'STORED\\r\\n' but got '%v'", out)
			}

			h1.verifyEmpty(t)
			h2.verifyEmpty(t)
		})
	})

	t.Run("Delete", func(t *testing.T) {
		// DELETE -> L1 OK
		t.Run("L1DeleteSuccess", func(t *testing.T) {
			h1 := &testHandler{
				errors: []error{nil},
			}
			h2 := &testHandler{}
			output := &bytes.Buffer{}

			l1only := orcas.L1OnlyCassandra(h1, h2, textprot.NewTextResponder(bufio.NewWriter(output)))

			err := l1only.Delete(common.DeleteRequest{
				Key:    []byte("key"),
				Opaque: 0,
				Quiet:  false,
			})
			if err != nil {
				t.Fatalf("Error should be nil, got %v", err)
			}

			out := string(output.Bytes())

			t.Logf(out)
			gold := "DELETED\r\n"

			if out != gold {
				t.Fatalf("Expected response '%v' but got '%v'", gold, out)
			}

			h1.verifyEmpty(t)
			h2.verifyEmpty(t)
		})

		// DELETE -> L1 Miss
		t.Run("L1DeleteMiss", func(t *testing.T) {
			h1 := &testHandler{
				errors: []error{common.ErrKeyNotFound},
			}
			h2 := &testHandler{}
			output := &bytes.Buffer{}

			l1only := orcas.L1OnlyCassandra(h1, h2, textprot.NewTextResponder(bufio.NewWriter(output)))

			err := l1only.Delete(common.DeleteRequest{
				Key:    []byte("key"),
				Opaque: 0,
				Quiet:  false,
			})
			if err != common.ErrKeyNotFound {
				t.Fatalf("Error should be %s, got %v", common.ErrKeyNotFound, err)
			}

			out := string(output.Bytes())

			if out != "" {
				t.Fatalf("Expected response is nil but got '%v'", out)
			}

			h1.verifyEmpty(t)
			h2.verifyEmpty(t)
		})

		// DELETE -> L1 Error
		t.Run("L1DeleteFail", func(t *testing.T) {
			h1 := &testHandler{
				errors: []error{common.ErrInternal},
			}
			h2 := &testHandler{}
			output := &bytes.Buffer{}

			l1only := orcas.L1OnlyCassandra(h1, h2, textprot.NewTextResponder(bufio.NewWriter(output)))

			err := l1only.Delete(common.DeleteRequest{
				Key:    []byte("key"),
				Opaque: 0,
				Quiet:  false,
			})
			if err != common.ErrInternal {
				t.Fatalf("Error should be %s, got %v", common.ErrInternal, err)
			}

			out := string(output.Bytes())

			t.Logf(out)
			gold := ""

			if out != gold {
				t.Fatalf("Expected response '%v' but got '%v'", gold, out)
			}

			h1.verifyEmpty(t)
			h2.verifyEmpty(t)
		})
	})

	t.Run("Replace", func(t *testing.T) {
		// REPLACE -> L1 OK
		t.Run("L1ReplaceSuccess", func(t *testing.T) {
			h1 := &testHandler{
				errors: []error{nil},
			}
			h2 := &testHandler{}
			output := &bytes.Buffer{}

			l1only := orcas.L1OnlyCassandra(h1, h2, textprot.NewTextResponder(bufio.NewWriter(output)))

			err := l1only.Replace(common.SetRequest{
				Key:    []byte("key"),
				Data:   []byte("value"),
				Opaque: 0,
				Quiet:  false,
			})
			if err != nil {
				t.Fatalf("Error should be nil, got %v", err)
			}

			out := string(output.Bytes())

			t.Logf(out)
			gold := "STORED\r\n"

			if out != gold {
				t.Fatalf("Expected response '%v' but got '%v'", gold, out)
			}

			h1.verifyEmpty(t)
			h2.verifyEmpty(t)
		})

		// REPLACE -> L1 MISS
		t.Run("L1ReplaceNotFound", func(t *testing.T) {
			h1 := &testHandler{
				errors: []error{common.ErrKeyNotFound},
			}
			h2 := &testHandler{}
			output := &bytes.Buffer{}

			l1only := orcas.L1OnlyCassandra(h1, h2, textprot.NewTextResponder(bufio.NewWriter(output)))

			err := l1only.Replace(common.SetRequest{
				Key:    []byte("key"),
				Data:   []byte("value"),
				Opaque: 0,
				Quiet:  false,
			})
			if err != common.ErrItemNotStored {
				t.Fatalf("Error should be %s, got %v", common.ErrItemNotStored, err)
			}

			out := string(output.Bytes())

			if out != "" {
				t.Fatalf("Expected response is nil but got '%v'", out)
			}

			h1.verifyEmpty(t)
			h2.verifyEmpty(t)
		})

		// REPLACE -> L1 Error
		t.Run("L1ReplaceFailure", func(t *testing.T) {
			h1 := &testHandler{
				errors: []error{common.ErrInternal},
			}
			h2 := &testHandler{}
			output := &bytes.Buffer{}

			l1only := orcas.L1OnlyCassandra(h1, h2, textprot.NewTextResponder(bufio.NewWriter(output)))

			err := l1only.Replace(common.SetRequest{
				Key:    []byte("key"),
				Data:   []byte("value"),
				Opaque: 0,
				Quiet:  false,
			})
			if err != common.ErrInternal {
				t.Fatalf("Error should be %s, got %v", common.ErrInternal, err)
			}

			out := string(output.Bytes())

			if out != "" {
				t.Fatalf("Expected response is nil but got '%v'", out)
			}

			h1.verifyEmpty(t)
			h2.verifyEmpty(t)
		})
	})

	// MISC & UNSUPPORTED COMMANDS
	t.Run("Touch", func(t *testing.T) {
		h1 := &testHandler{}
		h2 := &testHandler{}
		output := &bytes.Buffer{}
		l1only := orcas.L1OnlyCassandra(h1, h2, textprot.NewTextResponder(bufio.NewWriter(output)))

		err := l1only.Touch(common.TouchRequest{})
		if err != common.ErrUnknownCmd {
			t.Fatalf("Error should be %s, got %v", common.ErrUnknownCmd, err)
		}

		out := string(output.Bytes())
		gold := ""

		if out != gold {
			t.Fatalf("Expected response '%v' but got '%v'", gold, out)
		}

		h1.verifyEmpty(t)
		h2.verifyEmpty(t)
	})

	t.Run("Add", func(t *testing.T) {
		h1 := &testHandler{}
		h2 := &testHandler{}
		output := &bytes.Buffer{}
		l1only := orcas.L1OnlyCassandra(h1, h2, textprot.NewTextResponder(bufio.NewWriter(output)))

		err := l1only.Add(common.SetRequest{})
		if err != common.ErrUnknownCmd {
			t.Fatalf("Error should be %s, got %v", common.ErrUnknownCmd, err)
		}

		out := string(output.Bytes())
		gold := ""

		if out != gold {
			t.Fatalf("Expected response '%v' but got '%v'", gold, out)
		}

		h1.verifyEmpty(t)
		h2.verifyEmpty(t)
	})

	t.Run("Add", func(t *testing.T) {
		h1 := &testHandler{}
		h2 := &testHandler{}
		output := &bytes.Buffer{}
		l1only := orcas.L1OnlyCassandra(h1, h2, textprot.NewTextResponder(bufio.NewWriter(output)))

		err := l1only.Add(common.SetRequest{})
		if err != common.ErrUnknownCmd {
			t.Fatalf("Error should be %s, got %v", common.ErrUnknownCmd, err)
		}

		out := string(output.Bytes())
		gold := ""

		if out != gold {
			t.Fatalf("Expected response '%v' but got '%v'", gold, out)
		}

		h1.verifyEmpty(t)
		h2.verifyEmpty(t)
	})

	t.Run("Append", func(t *testing.T) {
		h1 := &testHandler{}
		h2 := &testHandler{}
		output := &bytes.Buffer{}
		l1only := orcas.L1OnlyCassandra(h1, h2, textprot.NewTextResponder(bufio.NewWriter(output)))

		err := l1only.Append(common.SetRequest{})
		if err != common.ErrUnknownCmd {
			t.Fatalf("Error should be %s, got %v", common.ErrUnknownCmd, err)
		}

		out := string(output.Bytes())
		gold := ""

		if out != gold {
			t.Fatalf("Expected response '%v' but got '%v'", gold, out)
		}

		h1.verifyEmpty(t)
		h2.verifyEmpty(t)
	})

	t.Run("Prepend", func(t *testing.T) {
		h1 := &testHandler{}
		h2 := &testHandler{}
		output := &bytes.Buffer{}
		l1only := orcas.L1OnlyCassandra(h1, h2, textprot.NewTextResponder(bufio.NewWriter(output)))

		err := l1only.Prepend(common.SetRequest{})
		if err != common.ErrUnknownCmd {
			t.Fatalf("Error should be %s, got %v", common.ErrUnknownCmd, err)
		}

		out := string(output.Bytes())
		gold := ""

		if out != gold {
			t.Fatalf("Expected response '%v' but got '%v'", gold, out)
		}

		h1.verifyEmpty(t)
		h2.verifyEmpty(t)
	})

	t.Run("GetE", func(t *testing.T) {
		h1 := &testHandler{}
		h2 := &testHandler{}
		output := &bytes.Buffer{}
		l1only := orcas.L1OnlyCassandra(h1, h2, textprot.NewTextResponder(bufio.NewWriter(output)))

		err := l1only.GetE(common.GetRequest{})
		if err != common.ErrUnknownCmd {
			t.Fatalf("Error should be %s, got %v", common.ErrUnknownCmd, err)
		}

		out := string(output.Bytes())
		gold := ""

		if out != gold {
			t.Fatalf("Expected response '%v' but got '%v'", gold, out)
		}

		h1.verifyEmpty(t)
		h2.verifyEmpty(t)
	})

	t.Run("Gat", func(t *testing.T) {
		h1 := &testHandler{}
		h2 := &testHandler{}
		output := &bytes.Buffer{}
		l1only := orcas.L1OnlyCassandra(h1, h2, textprot.NewTextResponder(bufio.NewWriter(output)))

		err := l1only.Gat(common.GATRequest{})
		if err != common.ErrUnknownCmd {
			t.Fatalf("Error should be %s, got %v", common.ErrUnknownCmd, err)
		}

		out := string(output.Bytes())
		gold := ""

		if out != gold {
			t.Fatalf("Expected response '%v' but got '%v'", gold, out)
		}

		h1.verifyEmpty(t)
		h2.verifyEmpty(t)
	})

	t.Run("Noop", func(t *testing.T) {
		h1 := &testHandler{}
		h2 := &testHandler{}
		output := &bytes.Buffer{}
		l1only := orcas.L1OnlyCassandra(h1, h2, textprot.NewTextResponder(bufio.NewWriter(output)))

		err := l1only.Noop(common.NoopRequest{})
		if err != nil {
			t.Fatalf("Error should be nil, got %v", err)
		}

		out := string(output.Bytes())
		gold := "Yep, it works.\r\n"

		if out != gold {
			t.Fatalf("Expected response '%v' but got '%v'", gold, out)
		}

		h1.verifyEmpty(t)
		h2.verifyEmpty(t)
	})

	t.Run("Version", func(t *testing.T) {
		h1 := &testHandler{}
		h2 := &testHandler{}
		output := &bytes.Buffer{}
		l1only := orcas.L1OnlyCassandra(h1, h2, textprot.NewTextResponder(bufio.NewWriter(output)))

		err := l1only.Version(common.VersionRequest{})
		if err != nil {
			t.Fatalf("Error should be nil, got %v", err)
		}

		out := string(output.Bytes())
		gold := fmt.Sprintf("VERSION %s\r\n", common.VersionString)

		if out != gold {
			t.Fatalf("Expected response '%v' but got '%v'", gold, out)
		}

		h1.verifyEmpty(t)
		h2.verifyEmpty(t)
	})

	t.Run("Quit", func(t *testing.T) {
		h1 := &testHandler{}
		h2 := &testHandler{}
		output := &bytes.Buffer{}
		l1only := orcas.L1OnlyCassandra(h1, h2, textprot.NewTextResponder(bufio.NewWriter(output)))

		err := l1only.Quit(common.QuitRequest{})
		if err != nil {
			t.Fatalf("Error should be nil, got %v", err)
		}

		out := string(output.Bytes())
		gold := "Bye\r\n"

		if out != gold {
			t.Fatalf("Expected response '%v' but got '%v'", gold, out)
		}

		h1.verifyEmpty(t)
		h2.verifyEmpty(t)
	})
}
