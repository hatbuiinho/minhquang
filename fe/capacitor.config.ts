import type { CapacitorConfig } from '@capacitor/cli';

const config: CapacitorConfig = {
	appId: 'vn.minhquang.personnel',
	appName: 'Nhân sự Minh Quang',
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
