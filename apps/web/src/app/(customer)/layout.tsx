import { CustomerNavbar } from "@/components/customer/header/navbar";
import { CustomerFooter } from "@/components/customer/footer/footer";

export default function CustomerLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="flex min-h-screen flex-col">
      <CustomerNavbar />
      <main className="flex-1">{children}</main>
      <CustomerFooter />
    </div>
  );
}
