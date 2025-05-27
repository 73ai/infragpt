import { SignIn } from "@clerk/clerk-react";

export default function LoginPage() {
  return (
    <div className="flex items-center justify-center h-screen">
      <SignIn signUpUrl="/signup" forceRedirectUrl={"/dashboard"} />
    </div>
  );
}