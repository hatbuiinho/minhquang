package vn.minhquang.personnel;

import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.os.Build;
import android.os.Bundle;
import com.getcapacitor.BridgeActivity;

public class MainActivity extends BridgeActivity {
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        createNotificationChannels();
    }

    private void createNotificationChannels() {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.O) {
            return;
        }

        NotificationManager manager = getSystemService(NotificationManager.class);
        if (manager == null) {
            return;
        }

        NotificationChannel normalChannel = new NotificationChannel(
                "general",
                getString(R.string.notification_channel_reminders_name),
                NotificationManager.IMPORTANCE_DEFAULT
        );
        normalChannel.setDescription(getString(R.string.notification_channel_reminders_description));

        NotificationChannel urgentChannel = new NotificationChannel(
                "urgent",
                getString(R.string.notification_channel_urgent_reminders_name),
                NotificationManager.IMPORTANCE_HIGH
        );
        urgentChannel.setDescription(getString(R.string.notification_channel_urgent_reminders_description));
        urgentChannel.enableVibration(true);

        manager.createNotificationChannel(normalChannel);
        manager.createNotificationChannel(urgentChannel);
    }
}
