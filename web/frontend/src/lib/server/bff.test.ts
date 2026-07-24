import { describe, it, expect, vi } from 'vitest';
import { forwardSessionCookie } from './bff';

function fakeCookies() {
	const set = vi.fn();
	const del = vi.fn();
	return { cookies: { set, delete: del } as any, set, del };
}

describe('forwardSessionCookie', () => {
	it('sets the session cookie from a Set-Cookie header', () => {
		const { cookies, set } = fakeCookies();
		const resp = new Response(null, { headers: { 'set-cookie': 'session=abc123; Path=/; HttpOnly' } });
		forwardSessionCookie(resp, cookies);
		expect(set).toHaveBeenCalledWith('session', 'abc123', expect.any(Object));
	});

	it('deletes the cookie when the value is empty', () => {
		const { cookies, del } = fakeCookies();
		const resp = new Response(null, { headers: { 'set-cookie': 'session=; Path=/; Max-Age=0' } });
		forwardSessionCookie(resp, cookies);
		expect(del).toHaveBeenCalledWith('session', { path: '/' });
	});

	it('does nothing when there is no Set-Cookie header', () => {
		const { cookies, set, del } = fakeCookies();
		forwardSessionCookie(new Response(null), cookies);
		expect(set).not.toHaveBeenCalled();
		expect(del).not.toHaveBeenCalled();
	});
});