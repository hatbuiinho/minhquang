import type { CapacitorConfig } from '@capacitor/cli';

const config: CapacitorConfig = {
	appId: 'com.ttpq.reminder',
	appName: 'Nhắc việc',
	webDir: 'build',
	android: {
		allowMixedContent: true
	},
	server: {
		androidScheme: 'https',
		cleartext: true
	},
	plugins: {
		CapacitorUpdater: {
			autoUpdate: false
		}
	}
};

export default config;
