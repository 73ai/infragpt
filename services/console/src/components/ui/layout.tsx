import { SidebarProvider } from "@/components/ui/sidebar"
import { ConsoleSidebar } from "@/components/console-sidebar"

const Layout = ({ children }: { children: React.ReactNode }) => {
  return (
    <SidebarProvider>
      <div className="flex min-h-screen w-full">
        <div className="sticky top-0 h-screen">
          <ConsoleSidebar />
        </div>
        <main className="flex-1 w-full min-w-0">
          <div className="flex flex-col w-full min-h-screen">
            {children}
          </div>
        </main>
      </div>
    </SidebarProvider>
  )
};

export default Layout;
