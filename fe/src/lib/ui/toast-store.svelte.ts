type ToastTone = 'success' | 'error' | 'info';

class ToastStore {
	message = $state('');
	tone = $state<ToastTone>('info');
	open = $state(false);
	private timeoutId: ReturnType<typeof setTimeout> | undefined;

	show(message: string, tone: ToastTone = 'info') {
		this.message = message;
		this.tone = tone;
		this.open = true;

		if (this.timeoutId) {
			clearTimeout(this.timeoutId);
		}
		this.timeoutId = setTimeout(() => this.close(), 3200);
	}

	success(message: string) {
		this.show(message, 'success');
	}

	error(message: string) {
		this.show(message, 'error');
	}

	info(message: string) {
		this.show(message, 'info');
	}

	close() {
		this.open = false;
		if (this.timeoutId) {
			clearTimeout(this.timeoutId);
			this.timeoutId = undefined;
		}
	}
}

export const toastStore = new ToastStore();
