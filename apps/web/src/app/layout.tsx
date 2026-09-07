import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: {
    template: "%s | SUN Booking Tours",
    default: "SUN Booking Tours - Explore Vietnam & Global Destinations",
  },
  description: "Book unforgettable tours, vacations, and travel experiences with SUN Booking Tours. Secure internet banking payment, verified reviews, and 24/7 customer support.",
  keywords: ["tour booking", "vietnam tours", "travel", "phu quoc", "sapa", "sun booking"],
  authors: [{ name: "SUN Booking Tours" }],
  metadataBase: new URL(process.env.NEXT_PUBLIC_SITE_URL || "http://localhost:3000"),
  openGraph: {
    type: "website",
    locale: "en_US",
    url: "https://sunbooking.com",
    siteName: "SUN Booking Tours",
    title: "SUN Booking Tours - Explore Amazing Destinations",
    description: "Book unforgettable tours and travel experiences with instant confirmation.",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className="min-h-screen antialiased flex flex-col font-sans">
        {children}
      </body>
    </html>
  );
}
