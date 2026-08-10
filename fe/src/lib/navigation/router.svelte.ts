import { parseRoute, type AppRoute, type MainRouteName } from './routes';

const initialRoute: AppRoute = {
	name: 'volunteers',
	path: '/volunteers',
	title: 'Huynh đệ công quả'
};

class AppRouter {
	current = $state<AppRoute>(initialRoute);
	private currentRoute = initialRoute;
	private initialized = false;
	private teardown?: () => void;

	init() {
		if (this.initialized) return;

		const route = parseRoute(window.location.pathname);
		this.setCurrent(route);
		if (window.location.pathname === '/') {
			this.replace(route.path);
		}

		const handlePopState = () => {
			this.setCurrent(parseRoute(window.location.pathname));
		};

		window.addEventListener('popstate', handlePopState);
		this.teardown = () => window.removeEventListener('popstate', handlePopState);
		this.initialized = true;
	}

	destroy() {
		this.teardown?.();
		this.teardown = undefined;
		this.initialized = false;
	}

	push(path: string) {
		const route = parseRoute(path);
		if (route.path === this.currentRoute.path) return;

		this.setCurrent(route);
		writeHistory(route.path, false);
	}

	replace(path: string) {
		const route = parseRoute(path);
		this.setCurrent(route);
		writeHistory(route.path, true);
	}

	back() {
		if (window.history.length > 1) {
			window.history.back();
			return;
		}

		this.replace('/volunteers');
	}

	openMain(name: MainRouteName) {
		const paths: Record<MainRouteName, string> = {
			volunteers: '/volunteers',
			departments: '/departments',
			users: '/users'
		};

		this.push(paths[name]);
	}

	private setCurrent(route: AppRoute) {
		this.current = route;
		this.currentRoute = route;
	}
}

export const router = new AppRouter();

function writeHistory(path: string, replaceState: boolean) {
	if (window.location.pathname === path) return;

	const state = { appRoute: true };
	if (replaceState) {
		window.history.replaceState(state, '', path);
		return;
	}

	window.history.pushState(state, '', path);
}
