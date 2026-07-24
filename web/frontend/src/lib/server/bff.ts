import { env } from '$env/dynamic/private';
import type { Cookies } from '@sveltejs/kit';

const BASE = env.BFF_BASE_URL ?? 'http://localhost:8081';

export async function bffFetch(path: string, cookieHeader: string, init: RequestInit = {}) {
	const headers = new Headers(init.headers);
	if (cookieHeader) headers.set('cookie', cookieHeader);
	return fetch(BASE + path, { ...init, headers });
}

export function forwardSessionCookie(resp: Response, cookies: Cookies) {
	const raw = resp.headers.get('set-cookie');
	if (!raw) return;
	const match = raw.match(/session=([^;]*)/);
	if (!match) return;
	const value = match[1];
	if (value === '') {
		cookies.delete('session', { path: '/' });
		return;
	}
	cookies.set('session', value, {
		path: '/',
		httpOnly: true,
		sameSite: 'lax',
		secure: env.SESSION_COOKIE_SECURE !== 'false',
		maxAge: 60 * 60 * 24 * 30
	});
}