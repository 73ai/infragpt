import "react";
import { SidebarTrigger } from "@/components/ui/sidebar";

const MainPage = () => {
  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="border-b">
        <div className="flex h-16 items-center px-4 gap-4">
          <SidebarTrigger />
          <h1 className="text-xl font-semibold">Overview</h1>
        </div>
      </div>
    </div>
  );
};

export default MainPage;
