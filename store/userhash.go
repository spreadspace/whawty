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
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/spreadspace/scryptauth.v2"
)

// fileExists returns whether the given file or directory exists or not
// this is from: stackoverflow.com/questions/10510691
func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// readHashStr returns the whole contents of the user hash file
func readHashStr(filename string) (string, string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return "", "", err
	}

	parts := strings.SplitN(string(data), ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("whawty.auth.store: hash file is invalid")
	}
	return parts[0], parts[1], nil
}

// IsFormatSupported checks if the format of the hash file is supported
func IsFormatSupported(filename string) (bool, error) {
	idStr, hashStr, err := readHashStr(filename)
	if err != nil {
		return false, err
	}
	if idStr != algoID {
		return false, fmt.Errorf("whawty.auth.store: hash file alogrithm ID '%s' is not supported", idStr)
	}

	ctxID, hash, salt, err := scryptauth.DecodeBase64(string(hashStr))
	if err != nil {
		return false, err
	}
	if ctxID == 0 || len(hash) == 0 || len(salt) == 0 {
		return false, nil
	}
	return true, nil
}

// UserHash is the representation of a single user hash file inside the store.
// Use NewUserHash to create it.
type UserHash struct {
	store *Dir
	user  string
}

// NewUserHash creates a new whawty.auth UserHash for user inside basedir.
func NewUserHash(store *Dir, user string) (u *UserHash) {
	u = &UserHash{}
	u.store = store
	u.user = user
	return
}

func (u *UserHash) getFilename(isAdmin bool) string {
	filename := filepath.Join(u.store.basedir, u.user)
	if isAdmin {
		return filename + adminExt
	}
	return filename + userExt
}

func (u *UserHash) writeHashStr(password string, isAdmin bool, flags int) error {
	ctx, ctxExists := u.store.Contexts[u.store.DefaultCtxID]
	if !ctxExists {
		return fmt.Errorf("whawty.auth.store: the store has no default context")
	}
	hash, salt, err := ctx.Gen([]byte(password))
	if err != nil {
		return err
	}

	file, err := os.OpenFile(u.getFilename(isAdmin), flags, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	hashStr := scryptauth.EncodeBase64(u.store.DefaultCtxID, hash, salt)
	_, err = io.WriteString(file, algoID+":"+hashStr+"\n") // TODO: retry if write was short??
	return err
}

// Add creates the hash file. It is an error if the user already exists.
func (u *UserHash) Add(password string, isAdmin bool) error {
	exists, _, err := u.Exists()
	if err != nil {
		return err
	} else if exists {
		return fmt.Errorf("whawty.auth.store: user '%s' already exists", u.user)
	}
	return u.writeHashStr(password, isAdmin, os.O_WRONLY|os.O_CREATE|os.O_EXCL)
}

// Update changes the password for user.
func (u *UserHash) Update(password string) error {
	exists, isAdmin, err := u.Exists()
	if err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("whawty.auth.store: user '%s' does not exist", u.user)
	}

	if ok, err := IsFormatSupported(u.getFilename(isAdmin)); err != nil || !ok {
		return fmt.Errorf("whawty.auth.store: won't overwrite unsupported hash format")
	}

	return u.writeHashStr(password, isAdmin, os.O_WRONLY|os.O_TRUNC)
}

// SetAdmin changes the admin status of user.
func (u *UserHash) SetAdmin(adminState bool) error {
	exists, isAdmin, err := u.Exists()
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("whawty.auth.store: user '%s' does not exist", u.user)
	}
	if isAdmin == adminState {
		return nil
	}

	oldname := filepath.Join(u.store.basedir, u.user)
	newname := oldname
	if adminState {
		oldname += userExt
		newname += adminExt
	} else {
		oldname += adminExt
		newname += userExt
	}
	return os.Rename(oldname, newname)
}

// Remove deletes hash file.
func (u *UserHash) Remove() {
	filename := filepath.Join(u.store.basedir, u.user)
	os.Remove(filename + adminExt)
	os.Remove(filename + userExt)
	return
}

// Exists checks if user exists. It also returns whether user is an admin. This returns true even if
// the user's hash file format is not supported
func (u *UserHash) Exists() (exists bool, isAdmin bool, err error) {
	filename := filepath.Join(u.store.basedir, u.user)
	var ok bool
	if ok, err = fileExists(filename + adminExt); err != nil {
		return
	} else if ok {
		return true, true, nil
	}
	isAdmin = false
	exists, err = fileExists(filename + userExt)
	return
}

// Authenticate checks the user password. It also returns whether user is an admin.
func (u *UserHash) Authenticate(password string) (isAuthenticated, isAdmin bool, err error) {
	var exists bool
	if exists, isAdmin, err = u.Exists(); err != nil {
		return
	} else if !exists {
		return false, false, fmt.Errorf("whawty.auth.store: user '%s' does not exist", u.user)
	}

	filename := filepath.Join(u.store.basedir, u.user)
	if isAdmin {
		filename += adminExt
	} else {
		filename += userExt
	}

	var idStr, hashStr string
	if idStr, hashStr, err = readHashStr(u.getFilename(isAdmin)); err != nil {
		return
	}
	if idStr != algoID {
		return false, false, fmt.Errorf("whawty.auth.store: user '%s' does not exist", u.user)
	}

	ctxID, hash, salt, err := scryptauth.DecodeBase64(hashStr)
	if err != nil {
		return false, false, err
	}
	ctx, ctxExists := u.store.Contexts[ctxID]
	if !ctxExists {
		return false, false, fmt.Errorf("whawty.auth.store: context ID '%d' is unknown", ctxID)
	}

	isAuthenticated, err = ctx.Check(hash, []byte(password), salt)
	return
}
