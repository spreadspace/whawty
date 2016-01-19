//
// Copyright (c) 2016 Christian Pointner <equinox@spreadspace.org>
//               2016 Markus Grüneis <gimpf@gimpf.org>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice, this
//   list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.
//
// * Neither the name of whawty.auth nor the names of its
//   contributors may be used to endorse or promote products derived from
//   this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//

package store

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

const (
	testBaseDir string = "test-store"
)

func TestAddRemoveUser(t *testing.T) {
	username := "test-addremove-user"
	password := "secret"

	s, _ := NewDir(testBaseDir)
	u := NewUserHash(s, username)

	if err := u.Add(password, false); err != nil {
		t.Fatal("unexpected error:", err)
	}
	if _, err := os.Stat(filepath.Join(testBaseDir, username+".user")); err != nil {
		t.Fatal("cannot read test user file after add:", err)
	}

	if err := u.Add(password, false); err == nil {
		t.Fatal("adding user a second time returned no error!")
	}

	u.Remove()
	if _, err := os.Stat(filepath.Join(testBaseDir, username+".user")); err == nil {
		t.Fatal("test user does still exist after remove")
	} else if !os.IsNotExist(err) {
		t.Fatal("unexpected error:", err)
	}
}

func TestAddRemoveAdmin(t *testing.T) {
	username := "test-addremove-admin"
	password := "secret"

	s, _ := NewDir(testBaseDir)
	u := NewUserHash(s, username)

	if err := u.Add(password, true); err != nil {
		t.Fatal("unexpected error:", err)
	}
	if _, err := os.Stat(filepath.Join(testBaseDir, username+".admin")); err != nil {
		t.Fatal("cannot read test user file after add:", err)
	}

	if err := u.Add(password, true); err == nil {
		t.Fatal("adding user a second time returned no error!")
	}

	u.Remove()
	if _, err := os.Stat(filepath.Join(testBaseDir, username+".admin")); err == nil {
		t.Fatal("test user does still exist after remove")
	} else if !os.IsNotExist(err) {
		t.Fatal("unexpected error:", err)
	}
}

func TestAddUserAdmin(t *testing.T) {
	username := "test-add-user-admin"
	password := "secret"

	s, _ := NewDir(testBaseDir)
	u := NewUserHash(s, username)

	if err := u.Add(password, false); err != nil {
		t.Fatal("unexpected error:", err)
	}
	defer u.Remove()

	if err := u.Add(password, true); err == nil {
		t.Fatal("re-adding existing user as admin shouldn't work")
	}
}

func TestAddAdminUser(t *testing.T) {
	username := "test-add-user-admin"
	password := "secret"

	s, _ := NewDir(testBaseDir)
	u := NewUserHash(s, username)

	if err := u.Add(password, true); err != nil {
		t.Fatal("unexpected error:", err)
	}
	defer u.Remove()

	if err := u.Add(password, false); err == nil {
		t.Fatal("re-adding existing user as normal user shouldn't work")
	}
}

func TestExistsUser(t *testing.T) {
	username := "test-exists-user"
	password := "secret"

	s, _ := NewDir(testBaseDir)
	u := NewUserHash(s, username)

	if exists, _, err := u.Exists(); err != nil {
		t.Fatal("unexpected error:", err)
	} else if exists {
		t.Fatal("hash file for test user shouldn't exist")
	}

	if err := u.Add(password, false); err != nil {
		t.Fatal("unexpected error:", err)
	}
	defer u.Remove()

	if exists, isAdmin, err := u.Exists(); err != nil {
		t.Fatal("unexpected error:", err)
	} else if !exists {
		t.Fatal("hash file for test user should exist")
	} else if isAdmin {
		t.Fatal("test user shouldn't be an admin")
	}
}

func TestExistsAdmin(t *testing.T) {
	username := "test-exists-admin"
	password := "secret"

	s, _ := NewDir(testBaseDir)
	u := NewUserHash(s, username)

	if err := u.Add(password, true); err != nil {
		t.Fatal("unexpected error:", err)
	}
	defer u.Remove()

	if exists, isAdmin, err := u.Exists(); err != nil {
		t.Fatal("unexpected error:", err)
	} else if !exists {
		t.Fatal("test user should exist")
	} else if !isAdmin {
		t.Fatal("test user should be an admin")
	}
}

func TestSetAdmin(t *testing.T) {
	username := "test-set-admin"
	password := "secret"

	s, _ := NewDir(testBaseDir)
	u := NewUserHash(s, username)

	if err := u.Add(password, false); err != nil {
		t.Fatal("unexpected error:", err)
	}
	defer u.Remove()

	if err := u.SetAdmin(true); err != nil {
		t.Fatal("unexpected error:", err)
	}

	if _, isAdmin, err := u.Exists(); err != nil {
		t.Fatal("unexpected error:", err)
	} else if !isAdmin {
		t.Fatal("test user should be an admin")
	}

	if err := u.SetAdmin(false); err != nil {
		t.Fatal("unexpected error:", err)
	}

	if err := u.SetAdmin(false); err != nil {
		t.Fatal("unexpected error:", err)
	}

	if _, isAdmin, err := u.Exists(); err != nil {
		t.Fatal("unexpected error:", err)
	} else if isAdmin {
		t.Fatal("test user shouldn't be an admin")
	}

	if err := u.SetAdmin(false); err != nil {
		t.Fatal("unexpected error:", err)
	}
}

func TestSetAdminNonExistent(t *testing.T) {
	username := "test-setadmin-nonexistent"

	s, _ := NewDir(testBaseDir)
	u := NewUserHash(s, username)

	if err := u.SetAdmin(true); err == nil {
		t.Fatal("setting admin on not exisiting user should be an error")
	}
}

func TestIsFormatSupported(t *testing.T) {
	username := "test-format-supported"
	password := "secret"
	username2 := "test-format-supported2"
	invalidStrings := []string{"", "hello", "hmac_sha256_scrypt:42:aGVsbG8=", "hmac_sha256_scrypt:0:aGVsbG8=:d29ybGQ=",
		"hmac_sha256_scrypt:214:aGVsbG8=:d29ybGQ=:d29ybGQ=", "hmac_sha256_scrypt:17:aGVsbG8=:d29ybGQ=:",
		"hmac_sha256_scrypt:23:aGVsbG8=:abcd$", "hmac_sha256_scrypt:12::aGVsbG8=", "hmac_sha256_scrypt::d29ybGQ=:aGVsbG8=",
		"hmac_sha256_scrypt:142:d29ybGQ=:", "hmac_sha1_scrypt:1:aGVsbG8=:d29ybGQ="}

	s, _ := NewDir(testBaseDir)
	u := NewUserHash(s, username)

	if err := u.Add(password, false); err != nil {
		t.Fatal("unexpected error:", err)
	}
	defer u.Remove()

	if ok, err := IsFormatSupported(filepath.Join(testBaseDir, username+".user")); err != nil {
		t.Fatal("unexpected error:", err)
	} else if !ok {
		t.Fatal("IsFormatSupported reported false negative")
	}

	filename := filepath.Join(testBaseDir, username2+".user")
	file, err := os.Create(filename)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	defer file.Close()
	defer os.Remove(filename)
	for _, invalid := range invalidStrings {
		if _, err := file.Seek(0, 0); err != nil {
			t.Fatal("unexpected error:", err)
		}
		if err := file.Truncate(0); err != nil {
			t.Fatal("unexpected error:", err)
		}
		if _, err := file.WriteString(invalid); err != nil {
			t.Fatal("unexpected error:", err)
		}

		if ok, err := IsFormatSupported(filename); err == nil && ok {
			t.Fatalf("IsFormatSupported reported false positive for '%s'", invalid)
		}
	}
}

func TestAuthenticate(t *testing.T) {
	username := "test-auth"
	password1 := "secret1"
	password2 := "secret2"

	s, _ := NewDir(testBaseDir)
	u := NewUserHash(s, username)

	if err := u.Add(password1, true); err != nil {
		t.Fatal("unexpected error:", err)
	}
	defer u.Remove()

	if isAuthOk, isAdmin, _ := u.Authenticate(password1); !isAuthOk {
		t.Fatal("authentication should succeed")
	} else if !isAdmin {
		t.Fatal("test user should be an admin")
	}

	if isAuthOk, isAdmin, _ := u.Authenticate(password2); isAuthOk {
		t.Fatal("authentication shouldn't succeed")
	} else if !isAdmin {
		t.Fatal("test user should be an admin")
	}
}

func TestAuthenticateNonExistent(t *testing.T) {
	username := "test-auth-nonexistent"
	password := "secret"

	s, _ := NewDir(testBaseDir)
	u := NewUserHash(s, username)

	if _, _, err := u.Authenticate(password); err == nil {
		t.Fatal("authenticating not exisiting user should be an error")
	}
}

func TestAuthenticateUnkownContext(t *testing.T) {
	username := "test-auth-unknown-ctx"
	password := "secret"
	hashStr := "hmac_sha256_scrypt:23:jYwMvYOTQ05_-MaOTwYuhDPPtGxt5wYHORLf93xDyQs=:RA-IO4_6GC2Qww4kFqMkstM5LejoPIWKHUPpTd0TU9w="

	filename := filepath.Join(testBaseDir, username+".user")
	file, err := os.Create(filename)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	defer file.Close()
	defer os.Remove(filename)
	if _, err := file.WriteString(hashStr); err != nil {
		t.Fatal("unexpected error:", err)
	}

	s, _ := NewDir(testBaseDir)
	u := NewUserHash(s, username)

	if _, _, err := u.Authenticate(password); err == nil {
		t.Fatal("authenticating a password which uses an unkown context should give an error")
	}
}

func TestAuthenticateInvalidHash(t *testing.T) {
	username := "test-auth-invalid-hash"
	password := "secret"
	hashStr := "hmac_sha256_scrypt:23:this is no salt:??"

	filename := filepath.Join(testBaseDir, username+".user")
	file, err := os.Create(filename)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	defer file.Close()
	defer os.Remove(filename)
	if _, err := file.WriteString(hashStr); err != nil {
		t.Fatal("unexpected error:", err)
	}

	s, _ := NewDir(testBaseDir)
	u := NewUserHash(s, username)

	if _, _, err := u.Authenticate(password); err == nil {
		t.Fatal("authenticating a password with an invalid hash string should give an error")
	}
}

func TestUpdateUser(t *testing.T) {
	username := "test-update-user"
	password1 := "secret"
	password2 := "moresecret"

	s, _ := NewDir(testBaseDir)
	u := NewUserHash(s, username)

	if err := u.Add(password1, false); err != nil {
		t.Fatal("unexpected error:", err)
	}
	defer u.Remove()

	if isAuthOk, _, _ := u.Authenticate(password1); !isAuthOk {
		t.Fatal("authentication should succeed")
	}
	if isAuthOk, _, _ := u.Authenticate(password2); isAuthOk {
		t.Fatal("authentication shouldn't succeed")
	}

	if err := u.Update(password2); err != nil {
		t.Fatal("unexpected error:", err)
	}
	if isAuthOk, _, _ := u.Authenticate(password1); isAuthOk {
		t.Fatal("authentication shouldn't succeed")
	}
	if isAuthOk, _, _ := u.Authenticate(password2); !isAuthOk {
		t.Fatal("authentication should succeed")
	}
}

func TestUpdateAdmin(t *testing.T) {
	username := "test-update-admin"
	password1 := "secret"
	password2 := "moresecret"

	s, _ := NewDir(testBaseDir)
	u := NewUserHash(s, username)

	if err := u.Add(password1, true); err != nil {
		t.Fatal("unexpected error:", err)
	}
	defer u.Remove()

	if isAuthOk, _, _ := u.Authenticate(password1); !isAuthOk {
		t.Fatal("authentication should succeed")
	}
	if isAuthOk, _, _ := u.Authenticate(password2); isAuthOk {
		t.Fatal("authentication shouldn't succeed")
	}

	if err := u.Update(password2); err != nil {
		t.Fatal("unexpected error:", err)
	}
	if isAuthOk, _, _ := u.Authenticate(password1); isAuthOk {
		t.Fatal("authentication shouldn't succeed")
	}
	if isAuthOk, _, _ := u.Authenticate(password2); !isAuthOk {
		t.Fatal("authentication should succeed")
	}
}

func TestUpdateNonExistent(t *testing.T) {
	username := "test-update-nonexistent"
	password := "secret"

	s, _ := NewDir(testBaseDir)
	u := NewUserHash(s, username)

	if err := u.Update(password); err == nil {
		t.Fatal("updating not exisiting user should be an error")
	}
}

func TestMain(m *testing.M) {
	if err := os.Mkdir(testBaseDir, 0755); err != nil {
		fmt.Println("Error creating store base directory:", err)
		os.Exit(-1)
	}

	ret := m.Run()

	if err := os.RemoveAll(testBaseDir); err != nil {
		fmt.Println("Error removing store base directory:", err)
	}
	os.Exit(ret)
}
