import { useEffect, useState } from "react";
import { observer } from "mobx-react-lite";
import { SidebarTrigger } from "@/components/ui/sidebar";

const MainPage = observer(() => {
  const [isOffline, setIsOffline] = useState(!navigator.onLine);

  // Listen for online/offline events
  useEffect(() => {
    const handleOnline = () => setIsOffline(false);
    const handleOffline = () => setIsOffline(true);

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);


  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="border-b">
        <div className="flex h-16 items-center px-4 gap-4">
        <SidebarTrigger />
          <h1 className="text-xl font-semibold">Overview</h1>
          {isOffline && (
            <span className="text-sm text-yellow-600 bg-yellow-100 px-2 py-1 rounded">
              Offline Mode
            </span>
          )}
        </div>
      </div>
    </div>
  );
});

export default MainPage; 