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
//
// This is based on simple-pam by T. Jameson Little licensed under MIT License.
// The source code of simple-pam can be found at:
//   https://github.com/beatgammit/simple-pam
// The complete license text can be found in the file LICENSE.simple-pam.
//

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define PAM_SM_AUTH

#include <security/pam_modules.h>

#define UNUSED(x) (void)(x)

/* internal functions */

#define WHAWTY_CONF_SILENT         0x01
#define WHAWTY_CONF_DEBUG          0x02
#define WHAWTY_CONF_USE_FIRST_PASS 0x04
#define WHAWTY_CONF_TRY_FIRST_PASS 0x08

typedef struct {
  int flags_;
  pam_handle_t* pamh_;
  const char* username_;
  char* password_;
} whawty_ctx_t;

int whawty_ctx_init(whawty_ctx_t* ctx, pam_handle_t *pamh, int flags, int argc, const char **argv)
{
  UNUSED(flags);
  UNUSED(argc);
  UNUSED(argv);

  ctx->flags_ = 0;
  ctx->pamh_ = pamh;
  ctx->password_ = NULL;

      // TODO: parse flags and arguments

  return pam_get_user(pamh, &(ctx->username_), NULL);
}

/* PAM Interface */

PAM_EXTERN int pam_sm_authenticate(pam_handle_t *pamh, int flags, int argc, const char **argv)
{
  whawty_ctx_t ctx;
  int ret = whawty_ctx_init(&ctx, pamh, flags, argc, argv);
  if(ret != PAM_SUCCESS)
    return ret;

  printf("whawty welcomes %s\n", ctx.username_);

  if (strcmp(ctx.username_, "equinox") != 0) {
    return PAM_AUTH_ERR;
  }

  return PAM_SUCCESS;
}

PAM_EXTERN int pam_sm_setcred(pam_handle_t *pamh, int flags, int argc, const char **argv)
{
  UNUSED(pamh);
  UNUSED(flags);
  UNUSED(argc);
  UNUSED(argv);
  return PAM_CRED_ERR;
}

/* static module data */
#ifdef PAM_STATIC

struct pam_module _pam_whawty_modstruct = {
    "pam_whawty",
    pam_sm_authenticate,
    pam_sm_setcred,
    NULL,
    NULL,
    NULL,
    NULL
};

#endif
