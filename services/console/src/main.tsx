import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App.tsx";
import "./index.css";
import { ClerkProvider } from "@clerk/clerk-react";

// Import your Publishable Key
const PUBLISHABLE_KEY = import.meta.env.VITE_CLERK_PUBLISHABLE_KEY;

if (!PUBLISHABLE_KEY) {
  throw new Error("Missing Publishable Key");
}

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <ClerkProvider
      publishableKey={PUBLISHABLE_KEY}
      signInUrl="/login"
      signUpUrl="/signup"
      appearance={{
        layout: {
          socialButtonsVariant: "iconButton",
        },
        elements: {
          formButtonPrimary: "bg-primary hover:bg-primary/90",
          card: "shadow-none border border-border",
        },
      }}
    >
      <App />
    </ClerkProvider>
  </React.StrictMode>,
);
