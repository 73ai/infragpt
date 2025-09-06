import { SignUp } from "@clerk/clerk-react";
import { useSearchParams } from "react-router-dom";

export default function SignUpPage() {
  const [searchParams] = useSearchParams();
  const redirectUrl = searchParams.get('redirect_url') || '/dashboard';

  const getDecodedRedirectUrl = (url: string): string => {
    try {
      return decodeURIComponent(url);
    } catch {
      return url;
    }
  };

  return (
    <div className="flex items-center justify-center h-screen">
        <SignUp 
          signInUrl="/login" 
          forceRedirectUrl={getDecodedRedirectUrl(redirectUrl)} 
        />
    </div>
  );
}