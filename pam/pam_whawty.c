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
#include <stdarg.h>

#define PAM_SM_AUTH

#include <security/pam_modules.h>
#include <security/pam_ext.h>
#include <syslog.h>
#include <security/_pam_macros.h>

#define UNUSED(x) (void)(x)

/* internal functions */

#define WHAWTY_CONF_SILENT         0x01
#define WHAWTY_CONF_DEBUG          0x02
#define WHAWTY_CONF_USE_FIRST_PASS 0x04
#define WHAWTY_CONF_TRY_FIRST_PASS 0x08
#define WHAWTY_CONF_NOT_SET_PASS   0x10

typedef struct {
  unsigned int flags_;
  pam_handle_t* pamh_;
  const char* username_;
  char* password_;
} whawty_ctx_t;

void PAM_FORMAT((printf, 3, 4)) _whawty_logf(whawty_ctx_t* ctx, int priority, const char* fmt, ...)
{
  if(ctx->flags_ & WHAWTY_CONF_SILENT)
    return;

  if(priority == LOG_DEBUG && !(ctx->flags_ & WHAWTY_CONF_DEBUG))
    return;

  va_list args;

  va_start(args, fmt);
  pam_vsyslog(ctx->pamh_, priority, fmt, args);
  va_end(args);
}

int _whawty_ctx_init(whawty_ctx_t* ctx, pam_handle_t *pamh, int flags, int argc, const char **argv)
{
  ctx->flags_ = 0;
  ctx->pamh_ = pamh;
  ctx->username_ = NULL;
  ctx->password_ = NULL;

  if(flags & PAM_SILENT)
    ctx->flags_ |= WHAWTY_CONF_SILENT;
  // flag PAM_DISALLOW_NULL_AUTHTOK is not applicable and will therefore be ignored

  int i;
  for(i = 0; i < argc; ++i) {
    if(!strcmp(argv[i], "debug"))
      ctx->flags_ |= WHAWTY_CONF_DEBUG;
    else if(!strcmp(argv[i], "try_first_pass"))
      ctx->flags_ |= WHAWTY_CONF_TRY_FIRST_PASS;
    else if(!strcmp(argv[i], "use_first_pass"))
      ctx->flags_ |= WHAWTY_CONF_USE_FIRST_PASS;
    else if(!strcmp(argv[i], "not_set_pass"))
      ctx->flags_ |= WHAWTY_CONF_NOT_SET_PASS;
    else
      _whawty_logf(ctx, LOG_WARNING, "ignoring unknown argument: %s", argv[i]);
  }

  int ret = pam_get_user(pamh, &(ctx->username_), NULL);
  if(ret == PAM_SUCCESS) {
    _whawty_logf(ctx, LOG_DEBUG, "successfully initialized [user=%s]", ctx->username_);
  } else {
    _whawty_logf(ctx, LOG_ERR, "pam_get_user() failed [%s]", pam_strerror(ctx->pamh_, ret));
  }
  return ret;
}

int _whawty_set_authtok_item(whawty_ctx_t* ctx) {
      // set PAM_AUTHTOK item to ctx->password_
  int ret = pam_set_item(ctx->pamh_, PAM_AUTHTOK, ctx->password_);
  if(ret != PAM_SUCCESS) {
    _pam_overwrite(ctx->password_);
    _pam_drop(ctx->password_);
    return ret;
  }
  return PAM_SUCCESS;
}

int _whawty_get_password(whawty_ctx_t* ctx)
{
  if(ctx->flags_ & WHAWTY_CONF_USE_FIRST_PASS || ctx->flags_ & WHAWTY_CONF_TRY_FIRST_PASS) {
        // fetch password from stack
    int ret = pam_get_item(ctx->pamh_, PAM_AUTHTOK, (const void**)(&(ctx->password_)));
    if(ret != PAM_SUCCESS) {
      _whawty_logf(ctx, LOG_ERR, "pam_get_item() returned an error reading the password [%s]", pam_strerror(ctx->pamh_, ret));
      return ret;
    }
    if(ctx->password_ != NULL) {
      _whawty_logf(ctx, LOG_DEBUG, "successfully fetched password [from stack]");
      return PAM_SUCCESS;
    }

    if(ctx->flags_ & WHAWTY_CONF_USE_FIRST_PASS) {
      _whawty_logf(ctx, LOG_ERR, "no password on stack and use_first_pass is set");
      return PAM_AUTHTOK_RECOVERY_ERR;
    }
  }

      // fetch password using the conversation function
  int ret = pam_prompt(ctx->pamh_, PAM_PROMPT_ECHO_OFF, &(ctx->password_), "Password: ");
  if(ret != PAM_SUCCESS) {
    if(ret == PAM_CONV_AGAIN) {
      _whawty_logf(ctx, LOG_DEBUG, "conversation function is not ready yet");
      return PAM_INCOMPLETE;
    }

    _whawty_logf(ctx, LOG_ERR, "pam_prompt() returned an error reading the password [%s]", pam_strerror(ctx->pamh_, ret));
    return ret;
  }
  if(ctx->password_ == NULL) {
    _whawty_logf(ctx, LOG_ERR, "conversation function returned no password");
    return PAM_AUTHTOK_RECOVERY_ERR;
  }

  _whawty_logf(ctx, LOG_DEBUG, "successfully fetched password [from conversation function]");

  if(!(ctx->flags_ & WHAWTY_CONF_NOT_SET_PASS)) {
    return _whawty_set_authtok_item(ctx);
  }

  return PAM_SUCCESS;
}

int _whawty_check_password(whawty_ctx_t* ctx)
{
      // TODO: open socket to whawty_auth and check password
  if (strcmp(ctx->username_, "equinox") != 0 || strcmp(ctx->password_, "test") != 0) {
    _whawty_logf(ctx, LOG_DEBUG, "authentication failure [user=%s]", ctx->username_);
    return PAM_AUTH_ERR;
  }

  _whawty_logf(ctx, LOG_NOTICE, "successfully authenticated [user=%s]", ctx->username_);
  return PAM_SUCCESS;
}

/* PAM Interface */

PAM_EXTERN int pam_sm_authenticate(pam_handle_t *pamh, int flags, int argc, const char **argv)
{
  whawty_ctx_t ctx;
  int ret = _whawty_ctx_init(&ctx, pamh, flags, argc, argv);
  if(ret != PAM_SUCCESS)
    return ret;

  ret = _whawty_get_password(&ctx);
  if(ret != PAM_SUCCESS)
    return ret;

  return _whawty_check_password(&ctx);
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
