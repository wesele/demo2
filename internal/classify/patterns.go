// Package classify provides client-side danger classification for shell commands.
package classify

import "regexp"

// Level represents the danger classification of a command.
type Level int

const (
	Safe    Level = iota // green
	Caution              // yellow
	Danger               // red
)

// String returns a human-readable label for the level.
func (l Level) String() string {
	switch l {
	case Danger:
		return "DANGER"
	case Caution:
		return "CAUTION"
	default:
		return "SAFE"
	}
}

// dangerPatterns matches commands that are potentially destructive or irreversible.
var dangerPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)rm\s+.*-[^\s]*r`),                 // recursive remove
	regexp.MustCompile(`(?i)rm\s+-[a-zA-Z]*r`),                // rm -rf style
	regexp.MustCompile(`(?i)\bdel\s+/[sf]`),                   // Windows force/recursive delete
	regexp.MustCompile(`(?i)\bmkfs\b`),                        // format filesystem
	regexp.MustCompile(`(?i)\bdd\s+if=`),                      // disk write
	regexp.MustCompile(`(?i)\bkill\s+-9\b`),                   // force kill
	regexp.MustCompile(`(?i)\bpkill\b`),                       // kill by name
	regexp.MustCompile(`(?i)\bkillall\b`),                     // kill all by name
	regexp.MustCompile(`:\(\)\s*\{.*:\|:.*\}`),                // fork bomb
	regexp.MustCompile(`(?i)>\s*/dev/sd`),                     // write to block device
	regexp.MustCompile(`(?i)\b(shutdown|reboot|halt|poweroff)\b`),
	regexp.MustCompile(`(?i)\bformat\s+[a-zA-Z]:`),            // Windows format drive
	regexp.MustCompile(`(?i)\bfdisk\b`),                       // partition table editor
	regexp.MustCompile(`(?i)\bparted\b`),                      // partition tool
	regexp.MustCompile(`(?i)\bwipefs\b`),                      // wipe filesystem signatures
	regexp.MustCompile(`(?i)\btruncate\s+.*-s\s+0`),           // zero a file
	regexp.MustCompile(`(?i)\bdrop\s+(database|table)\b`),     // SQL destructive
	regexp.MustCompile(`(?i)\bStop-Computer\b`),               // PowerShell shutdown
	regexp.MustCompile(`(?i)\bRemove-Item\s+.*-Recurse`),      // PowerShell recursive delete
}

// cautionPatterns matches commands that create or modify files/state.
var cautionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bmkdir\b`),
	regexp.MustCompile(`(?i)\btouch\b`),
	regexp.MustCompile(`(?i)\bcp\b`),
	regexp.MustCompile(`(?i)\bmv\b`),
	regexp.MustCompile(`(?i)\bln\b`),
	regexp.MustCompile(`(?i)\bchmod\b`),
	regexp.MustCompile(`(?i)\bchown\b`),
	regexp.MustCompile(`(?i)\bSet-Content\b`),
	regexp.MustCompile(`(?i)\bNew-Item\b`),
	regexp.MustCompile(`(?i)\bCopy-Item\b`),
	regexp.MustCompile(`(?i)\bMove-Item\b`),
	regexp.MustCompile(`(?i)\bcurl\b.*-[oO]\b`),
	regexp.MustCompile(`(?i)\bwget\b`),
	regexp.MustCompile(`(?i)\bpip\s+install\b`),
	regexp.MustCompile(`(?i)\bnpm\s+install\b`),
	regexp.MustCompile(`(?i)\byarn\s+add\b`),
	regexp.MustCompile(`(?i)\bapt(-get)?\s+(install|remove|purge)\b`),
	regexp.MustCompile(`(?i)\byum\s+(install|remove)\b`),
	regexp.MustCompile(`(?i)\bbrew\s+(install|uninstall|remove)\b`),
	regexp.MustCompile(`(?i)\bsystemctl\s+(start|stop|restart|enable|disable)\b`),
	regexp.MustCompile(`(?i)\bservice\s+\S+\s+(start|stop|restart)\b`),
	regexp.MustCompile(`(?i)\bdocker\s+(run|stop|rm|rmi|build)\b`),
	regexp.MustCompile(`(?i)\bkubectl\s+(apply|delete|scale)\b`),
	regexp.MustCompile(`(?i)\bgit\s+(push|reset|rebase|force)\b`),
	regexp.MustCompile(`(?i)\bsed\s+.*-i\b`),                 // in-place sed edit
	regexp.MustCompile(`(?i)\bawk\b.*>\s*\S+`),               // awk redirecting output
}
