export type MainRouteName = 'events' | 'calendar' | 'reminders' | 'settings';
export type DetailRouteName = 'event-detail' | 'event-new' | 'event-edit';
export type RouteName = MainRouteName | DetailRouteName;

export type AppRoute = {
	name: RouteName;
	path: string;
	title: string;
	eventId?: string;
};

export type BottomNavItem = {
	name: MainRouteName;
	path: string;
	label: string;
	icon: string;
};

export const bottomNavItems: BottomNavItem[] = [
	{ name: 'events', path: '/events', label: 'Sự kiện', icon: 'icon-[lucide--list-todo]' },
	{ name: 'calendar', path: '/calendar', label: 'Lịch', icon: 'icon-[lucide--calendar-days]' },
	{ name: 'reminders', path: '/reminders', label: 'Nhắc hẹn', icon: 'icon-[lucide--bell-ring]' },
	{ name: 'settings', path: '/settings', label: 'Cài đặt', icon: 'icon-[lucide--settings]' }
];

export function parseRoute(pathname: string): AppRoute {
	const path = normalizePath(pathname);

	if (path === '/' || path === '/events') {
		return { name: 'events', path: '/events', title: 'Sự kiện' };
	}
	if (path === '/events/new') {
		return { name: 'event-new', path, title: 'Tạo sự kiện' };
	}

	const eventEditMatch = /^\/events\/([^/]+)\/edit$/.exec(path);
	if (eventEditMatch) {
		return {
			name: 'event-edit',
			path,
			title: 'Sửa sự kiện',
			eventId: decodeURIComponent(eventEditMatch[1])
		};
	}

	const eventMatch = /^\/events\/([^/]+)$/.exec(path);
	if (eventMatch) {
		return {
			name: 'event-detail',
			path,
			title: 'Chi tiết sự kiện',
			eventId: decodeURIComponent(eventMatch[1])
		};
	}

	if (path === '/calendar') {
		return { name: 'calendar', path, title: 'Lịch' };
	}
	if (path === '/reminders') {
		return { name: 'reminders', path, title: 'Nhắc hẹn' };
	}
	if (path === '/settings') {
		return { name: 'settings', path, title: 'Cài đặt' };
	}

	return { name: 'events', path: '/events', title: 'Sự kiện' };
}

export function mainRouteFor(route: AppRoute): MainRouteName {
	if (route.name === 'event-detail' || route.name === 'event-new' || route.name === 'event-edit') {
		return 'events';
	}

	return route.name;
}

function normalizePath(pathname: string): string {
	if (pathname.length > 1 && pathname.endsWith('/')) {
		return pathname.slice(0, -1);
	}
	return pathname;
}
