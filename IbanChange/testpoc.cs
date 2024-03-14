using System;
using System.ServiceProcess;
using System.Text.RegularExpressions;
using System.Threading;
using System.Windows.Forms;

public class ClipboardMonitorService : ServiceBase
{
    private Timer timer;

    public ClipboardMonitorService()
    {
        this.ServiceName = "ClipboardMonitorService";
        this.CanStop = true;
    }

    protected override void OnStart(string[] args)
    {
        timer = new Timer(CheckClipboard, null, TimeSpan.Zero, TimeSpan.FromSeconds(1));
    }

    protected override void OnStop()
    {
        timer.Dispose();
    }

    private void CheckClipboard(object state)
    {
        if (Clipboard.ContainsText())
        {
            string clipboardText = Clipboard.GetText();

            // Check if clipboard text matches the regex pattern
            if (Regex.IsMatch(clipboardText, @"^(?:[A-Z]{2})(?:(?![a-zA-Z0-9]{2}\b)\w){2,30}$", RegexOptions.IgnoreCase))
            {
                // Replace the pattern with "test"
                clipboardText = Regex.Replace(clipboardText, @"^(?:[A-Z]{2})(?:(?![a-zA-Z0-9]{2}\b)\w){2,30}$", "test", RegexOptions.IgnoreCase);

                Console.WriteLine("Text from clipboard:");
                Console.WriteLine(clipboardText);
            }
            else
            {
                Console.WriteLine("Clipboard text does not match the pattern. Ignoring...");
            }
        }
        else
        {
            Console.WriteLine("No text found in the clipboard.");
        }
    }

    public static void Main()
    {
        ServiceBase.Run(new ClipboardMonitorService());
    }
}
