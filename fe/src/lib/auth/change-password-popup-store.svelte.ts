class ChangePasswordPopupStore {
	open = $state(false);

	show() {
		this.open = true;
	}

	close() {
		this.open = false;
	}
}

export const changePasswordPopupStore = new ChangePasswordPopupStore();
