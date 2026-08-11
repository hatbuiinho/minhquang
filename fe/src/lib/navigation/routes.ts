export type MainRouteName = 'volunteers' | 'departments' | 'users';
export type RouteName = MainRouteName | 'volunteer-detail' | 'volunteer-new' | 'volunteer-edit';

export type AppRoute = {
	name: RouteName;
	path: string;
	title: string;
	volunteerId?: string;
};

export const bottomNavItems = [
	{
		name: 'volunteers' as const,
		path: '/volunteers',
		label: 'Công quả',
		icon: 'icon-[lucide--users]'
	},
	{
		name: 'departments' as const,
		path: '/departments',
		label: 'Phân ban',
		icon: 'icon-[lucide--layout-list]'
	},
	{ name: 'users' as const, path: '/users', label: 'Tài khoản', icon: 'icon-[lucide--shield-user]' }
];

export function parseRoute(pathname: string): AppRoute {
	const path = pathname.length > 1 && pathname.endsWith('/') ? pathname.slice(0, -1) : pathname;
	if (path === '/' || path === '/volunteers')
		return { name: 'volunteers', path: '/volunteers', title: 'Huynh đệ công quả' };
	if (path === '/volunteers/new') return { name: 'volunteer-new', path, title: 'Thêm Huynh đệ' };
	const edit = /^\/volunteers\/([^/]+)\/edit$/.exec(path);
	if (edit)
		return {
			name: 'volunteer-edit',
			path,
			title: 'Sửa Huynh đệ',
			volunteerId: decodeURIComponent(edit[1])
		};
	const detail = /^\/volunteers\/([^/]+)$/.exec(path);
	if (detail)
		return {
			name: 'volunteer-detail',
			path,
			title: 'Chi tiết Huynh đệ',
			volunteerId: decodeURIComponent(detail[1])
		};
	if (path === '/users') return { name: 'users', path, title: 'Tài khoản quản trị' };
	if (path === '/departments') return { name: 'departments', path, title: 'Quản lý phân ban' };
	return { name: 'volunteers', path: '/volunteers', title: 'Huynh đệ công quả' };
}

export function mainRouteFor(route: AppRoute): MainRouteName {
	if (route.name === 'users' || route.name === 'departments') return route.name;
	return 'volunteers';
}
