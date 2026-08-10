type PopupTone = 'default' | 'danger';

type ConfirmOptions = {
	title: string;
	message: string;
	confirmLabel?: string;
	cancelLabel?: string;
	tone?: PopupTone;
};

class PopupStore {
	open = $state(false);
	title = $state('');
	message = $state('');
	confirmLabel = $state('Đồng ý');
	cancelLabel = $state('Huỷ');
	tone = $state<PopupTone>('default');
	private resolveConfirm?: (confirmed: boolean) => void;

	confirm(options: ConfirmOptions): Promise<boolean> {
		this.resolve(false);

		this.title = options.title;
		this.message = options.message;
		this.confirmLabel = options.confirmLabel ?? 'Đồng ý';
		this.cancelLabel = options.cancelLabel ?? 'Huỷ';
		this.tone = options.tone ?? 'default';
		this.open = true;

		return new Promise((resolve) => {
			this.resolveConfirm = resolve;
		});
	}

	cancel() {
		this.resolve(false);
	}

	accept() {
		this.resolve(true);
	}

	private resolve(confirmed: boolean) {
		if (!this.resolveConfirm) {
			this.open = false;
			return;
		}

		const resolve = this.resolveConfirm;
		this.resolveConfirm = undefined;
		this.open = false;
		resolve(confirmed);
	}
}

export const popupStore = new PopupStore();
