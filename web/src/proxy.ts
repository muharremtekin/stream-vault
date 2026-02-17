import { NextRequest, NextResponse } from 'next/server';

const TOKEN_COOKIE = 'sv-access-token';

const PUBLIC_ROUTES = ['/', '/login', '/register', '/profile-select'];

const AUTH_ONLY_ROUTES = ['/login', '/register'];

export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const token = request.cookies.get(TOKEN_COOKIE)?.value;

  if (token && AUTH_ONLY_ROUTES.includes(pathname)) {
    return NextResponse.redirect(new URL('/browse', request.url));
  }

  if (token && pathname === '/') {
    return NextResponse.redirect(new URL('/browse', request.url));
  }

  const isPublic = PUBLIC_ROUTES.includes(pathname);
  if (!token && !isPublic) {
    return NextResponse.redirect(new URL('/login', request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico).*)'],
};
