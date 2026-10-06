package shell

// ZshInit installs the shell-local command capture hook and an r wrapper that
// exposes the captured command only to the r child process.
const ZshInit = `autoload -Uz add-zsh-hook

typeset -g _R_LAST_COMMAND=''

_r_capture_preexec() {
  local -a command_words
  command_words=("${(z)1}")
  if (( ${#command_words} > 0 )) && [[ "${command_words[1]}" == 'r' ]]; then
    return
  fi
  typeset -g _R_LAST_COMMAND="$1"
}

r() {
  R_LAST_COMMAND="${_R_LAST_COMMAND-}" command r "$@"
}

add-zsh-hook -d preexec _r_capture_preexec 2>/dev/null
add-zsh-hook preexec _r_capture_preexec
`
