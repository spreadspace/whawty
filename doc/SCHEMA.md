# whawty.auth storage schema

The whawty.auth store consists of a single directory which contains all
usernames and password hashes in flat files. Password files are named
after the user. The file extension is used to distinguish between admin
and normal user.

    /path/to/whawty/auth/base
      adminuser.admin       ; password hash for user adminuser which is an admin
      gimpf.admin           ; there are multiple admins allowed
      equinox.user          ; password hash for user equinox
      fredl.user            ; password hash for user fredl

The difference between admins and normal users is that admin users are
allowed to add new users. Also granting admin privileges to normal users
may only be done by admins. An admin is also allowed to change any password.
Normal users may only update their own password.
However this mechanism is not intended as a replacement for a real authorization
database but should only be used by management interfaces of a whawty.auth agent.
You may want to take a look at [whawty.groups](https://github.com/whawty/groups)
if you need an authorization database.

The directory may also contains a `.tmp` directory: agents that update
the whawty.auth base must perform file modifications atomically by
writing new files, with a random name, in that directory, then
atomically moving them to their final destination.  As such, `.tmp`
should be backed by the same filesystem as the whawty.auth base.

The directory must not contain any other files. A valid whawty.auth base
directory contains at least one admin file which uses a supported hashing
format.
Furthermore a directory may contain only one hash file per user.
If this conditions are not met the agent has to exit with an error.

If a whawty.auth agent doesn't support the hashing format of a file it has
to act according to the following rules:

- **on authenticate:** ignore the file and act as if the user does not exist
- **on add:** report an error (user exists)
- **on update:** report an error (won't overwrite unsupported formats)
- **on delete:** delete the file (a warning may be shown)

Usernames must only contain the following characters: `[-_.@A-Za-z0-9]`

A whawty.auth agent may upgrade the hashing algorithm to an other (newer/stronger)
format during authentication.
However if an agent supports this it must be possible to disable upgrades.


## File Format

The first line of the file contains the password hash which has the following format:

    <format-identifier>:<last-change>:<format specific string>

`format-identifier` is a unique identifier for the hashing format. This id must
not include a `:`. `last-change` is a UNIX time stamp and represents the last
date/time when the password has been modified.

The rest of the file (first line excluded) is reserved for auxiliary data.


## Hashing formats

For now the only supported algorithm is scrypt inside hmac-sha256 which has the
following structure:

    hmac_sha256_scrypt:<last-change>:ctxID:base64(salt):base64(hash)

`hmac_sha256_scrypt` is the identifier for this algorithm, `ctxID` is an
identifier for a set of parameters which must be stored outside of the base
directory. A whawty.auth agent should support multiple parameter-sets to allow
soft upgrades of passwords. This algorithm needs the following parameters:

    server_key: the key for the hmac-sha256
    pw_cost:    (1<<pw_cost) forms the scrypt parameter N
    r:          the scrypt parameter r
    p:          the scrypt parameter p

`salt` is a random number with 256bits, `hash` is the output of the following
function:

    hmac_sha256(scrypt(user_password, salt, N, r, p), server_key)


## Auxiliary Data

The auxiliary data associated to a user is stored in the user's file, after the
first line.  It is organized as key value pairs.

A file may contain 0 or more lines of auxiliary data,
each of which starts with a unique identifier:

    <identifier>: <base64(data)>

`identifier` must not contain a `:` and must be unique (see table below). For now
aux-data is only used for 2/multi-factor authentication schemes but might be used
for other purposes as well. The values are base64 encoded and besides this encoding
shouldn't be mangled with by a whawty.auth agent.


| Identifier | Description                                  |
|------------|----------------------------------------------|
| `u2f`      | FIDO Universal 2nd Factor Token              |
| `totp`     | Time-based One-Time Password Token (RFC6238) |
