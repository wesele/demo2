package classify

import "testing"

func TestDangerPatterns(t *testing.T) {
	dangerCmds := []string{
		"rm -rf /tmp/logs",
		"rm -r /var/data",
		"del /f /s /q C:\\Temp",
		"mkfs.ext4 /dev/sdb",
		"dd if=/dev/zero of=/dev/sda",
		"kill -9 1234",
		"pkill nginx",
		"killall java",
		"shutdown -h now",
		"reboot",
		"halt",
		"poweroff",
		"format C:",
		"fdisk /dev/sda",
		"Stop-Computer",
		"Remove-Item -Recurse -Force ./dist",
	}
	for _, cmd := range dangerCmds {
		got := Classify(cmd)
		if got != Danger {
			t.Errorf("Classify(%q) = %v, want Danger", cmd, got)
		}
	}
}

func TestCautionPatterns(t *testing.T) {
	cautionCmds := []string{
		"mkdir /tmp/test",
		"touch file.txt",
		"cp src dst",
		"mv file.txt /tmp/",
		"chmod 755 script.sh",
		"chown user:group file",
		"Set-Content -Path file.txt -Value data",
		"New-Item -ItemType File -Path test.txt",
		"Copy-Item src dst",
		"Move-Item src dst",
		"curl -o output.txt https://example.com",
		"wget https://example.com/file.zip",
		"pip install requests",
		"npm install express",
		"apt install git",
		"brew install jq",
		"systemctl restart nginx",
		"docker run ubuntu",
		"kubectl delete pod mypod",
		"git push origin main",
		"sed -i 's/foo/bar/g' file.txt",
	}
	for _, cmd := range cautionCmds {
		got := Classify(cmd)
		if got != Caution {
			t.Errorf("Classify(%q) = %v, want Caution", cmd, got)
		}
	}
}

func TestSafePatterns(t *testing.T) {
	safeCmds := []string{
		"ls -la",
		"cat file.txt",
		"echo hello",
		"pwd",
		"whoami",
		"date",
		"ps aux",
		"Get-ChildItem",
		"grep pattern file.txt",
		"find /tmp -name '*.log'",
		"df -h",
		"top",
		"curl https://example.com",  // curl without -o is safe
	}
	for _, cmd := range safeCmds {
		got := Classify(cmd)
		if got != Safe {
			t.Errorf("Classify(%q) = %v, want Safe", cmd, got)
		}
	}
}

func TestLevelString(t *testing.T) {
	if Safe.String() != "SAFE" {
		t.Errorf("Safe.String() = %q", Safe.String())
	}
	if Caution.String() != "CAUTION" {
		t.Errorf("Caution.String() = %q", Caution.String())
	}
	if Danger.String() != "DANGER" {
		t.Errorf("Danger.String() = %q", Danger.String())
	}
}
