import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter, Routes, Route } from "react-router";
import "@/globals.css";
import { HomePage } from "@/pages/HomePage";
import { Layout } from "@/pages/Layout";
import { ProtectedRoute } from "@/components/ProtectedRoute";
import { RegisterPage } from "@/pages/RegisterPage";
import { AuthProvider } from "@/components/Providers/AuthProvider";
import { LoginPage } from "@/pages/LoginPage";
import { VerifyPage } from "@/pages/VerifyPage";
import { ResetPasswordPage } from "@/pages/ResetPasswordPage";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route element={<Layout />}>
            <Route path="login" element={<LoginPage />} />
            <Route path="reset" element={<ResetPasswordPage />} />
            <Route path="register" element={<RegisterPage />} />
            <Route path="verify" element={<VerifyPage />} />

            <Route element={<ProtectedRoute />}>
              <Route index element={<HomePage />} />
            </Route>
          </Route>
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  </StrictMode>,
);
