import { SignIn } from "@clerk/clerk-react";
import { useSearchParams } from "react-router-dom";

export default function LoginPage() {
  const [searchParams] = useSearchParams();
  const redirectUrl = searchParams.get('redirect_url') || '/dashboard';

  return (
    <div className="flex items-center justify-center h-screen">
      <SignIn 
        signUpUrl="/signup" 
        forceRedirectUrl={decodeURIComponent(redirectUrl)} 
      />
    </div>
  );
}