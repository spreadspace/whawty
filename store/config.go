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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/spreadspace/scryptauth.v2"
)

type cfgCtx struct {
	ID            uint   `json:"id"`
	HmacKeyBase64 string `json:"hmackey"`
	PwCost        uint   `json:"pwcost"`
	R             int    `json:"r"`
	P             int    `json:"p"`
}

type config struct {
	BaseDir    string   `json:"basedir"`
	DefaultCtx uint     `json:"defaultctx"`
	Contexts   []cfgCtx `json:"contexts"`
}

func readConfig(configfile string) (*config, error) {
	file, err := os.Open(configfile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	jsondata, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	c := &config{}
	if jsonerr := json.Unmarshal(jsondata, c); jsonerr != nil {
		return nil, fmt.Errorf("Error parsing config file: %s", jsonerr)
	}
	return c, nil
}

func (d *Dir) fromConfig(configfile string) error {
	c, err := readConfig(configfile)
	if err != nil {
		return err
	}
	if c.BaseDir == "" {
		return fmt.Errorf("Error: config file does not contain a base directory")
	}
	d.basedir = c.BaseDir
	for _, ctx := range c.Contexts {
		if ctx.ID == 0 {
			return fmt.Errorf("Error: context ID 0 is not allowed")
		}
		hk, err := base64.URLEncoding.DecodeString(ctx.HmacKeyBase64)
		if err != nil {
			return fmt.Errorf("Error: can't decode HMAC Key for context ID %d: %s", ctx.ID, err)
		}
		if len(hk) != scryptauth.KeyLength {
			return fmt.Errorf("Error: HMAC Key for context ID %d has invalid length %d != %d", ctx.ID, scryptauth.KeyLength, len(hk))
		}

		sactx, err := scryptauth.New(ctx.PwCost, hk)
		if err != nil {
			return err
		}
		if ctx.R > 0 {
			sactx.R = ctx.R
		}
		if ctx.P > 0 {
			sactx.P = ctx.P
		}
		d.Contexts[ctx.ID] = sactx
	}
	if c.DefaultCtx == 0 {
		if len(d.Contexts) != 0 {
			return fmt.Errorf("Error: no default context")
		}
	} else if _, exists := d.Contexts[c.DefaultCtx]; !exists {
		return fmt.Errorf("Error: invalid default context %d", c.DefaultCtx)
	}
	d.DefaultCtxID = c.DefaultCtx
	return nil
}
