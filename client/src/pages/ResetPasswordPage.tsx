import { Button } from "@/components/ui/Button";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm, type SubmitHandler } from "react-hook-form";
import { z } from "zod";
import { api } from "@/api/axiosInstance";
import { Loader } from "@/components/ui/Loader";
import { Form } from "@/components/Form/Form";
import { Input } from "@/components/Form/Input";
import { useState } from "react";
import { useSearchParams } from "react-router";
import { useNavigate } from "react-router";

const emailSchema = z.object({
  email: z
    .string({ message: "Email is required" })
    .email({ message: "Email is invalid" }),
});
const passwordSchema = z
  .object({
    password: z
      .string({ message: "Password is required" })
      .min(6, { message: "Password is too short (6 letters min.)" }),
    confirmPassword: z.string(),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords don't match",
    path: ["confirmPassword"],
  });

type EmailFormFields = z.infer<typeof emailSchema>;
type PasswordFormFields = z.infer<typeof passwordSchema>;

export const ResetPasswordPage = () => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  const {
    register: registerEmail,
    handleSubmit: handleSubmitEmail,
    formState: {
      isSubmitting: isSubmittingEmail,
      errors: errorsEmail,
      isValid: isValidEmail,
    },
  } = useForm<EmailFormFields>({
    resolver: zodResolver(emailSchema),
    mode: "onChange",
  });

  const {
    register: registerPassword,
    handleSubmit: handleSubmitPassword,
    setError: setErrorPassword,
    formState: {
      isSubmitting: isSubmittingPassword,
      errors: errorsPassword,
      isValid: isValidPassword,
    },
  } = useForm<PasswordFormFields>({
    resolver: zodResolver(passwordSchema),
    mode: "onChange",
  });

  const [isSent, setIsSent] = useState(false);

  const onSubmitEmail: SubmitHandler<EmailFormFields> = async (data) => {
    await api.post("/v1/request-password-reset", data);
    setIsSent(true);
  };

  const onSubmitPassword: SubmitHandler<PasswordFormFields> = async (data) => {
    const reqData = {
      password: data.password,
    };

    try {
      await api.patch(
        `/v1/reset-password?token=${searchParams.get("token")}`,
        reqData,
      );
      navigate("/login");
    } catch {
      setErrorPassword("root", { message: "Something went wrong" });
    }
  };

  return (
    <main className="flex h-full w-[400px] flex-col items-center justify-center gap-6">
      {searchParams.has("token") ? (
        <>
          <div className="flex flex-col items-center gap-3 text-center">
            <h1 className="text-[1.25rem]">Reset your password</h1>
            <p className="text-smoke-300 text-[0.875rem]">
              Enter your new password.
            </p>
          </div>
          <Form
            formError={errorsPassword.root?.message}
            onSubmit={handleSubmitPassword(onSubmitPassword)}
          >
            <Input
              {...registerPassword("password")}
              name="password"
              type="password"
              placeholder="New password"
              inputError={errorsPassword.password?.message}
              first
            />
            <Input
              {...registerPassword("confirmPassword")}
              name="confirmPassword"
              type="password"
              placeholder="Confirm password"
              inputError={errorsPassword.confirmPassword?.message}
            />
            <Button
              disabled={isSubmittingPassword || !isValidPassword}
              type="submit"
              text={isSubmittingPassword ? "" : "Reset password"}
              size="md"
              className="w-full"
            >
              {isSubmittingPassword ? <Loader width={24} height={24} /> : <></>}
            </Button>
          </Form>
        </>
      ) : isSent ? (
        <div className="flex flex-col items-center gap-4 text-center">
          <h1 className="text-[1.25rem]">We've sent you a reset link</h1>
          <p className="text-smoke-300 text-[0.875rem]">
            Please check your{" "}
            <a
              href="https://mail.google.com"
              className="underline-hover text-smoke-100 hover:text-accent relative transition-all"
            >
              email
            </a>
            . You may close this page now.
          </p>
        </div>
      ) : (
        <>
          <div className="flex flex-col items-center gap-3 text-center">
            <h1 className="text-[1.25rem]">Reset your password</h1>
            <p className="text-smoke-300 text-[0.875rem]">
              Enter your email address to reset password.
            </p>
          </div>
          <Form
            formError={errorsEmail.root?.message}
            onSubmit={handleSubmitEmail(onSubmitEmail)}
          >
            <Input
              {...registerEmail("email")}
              name="email"
              type="text"
              placeholder="Email"
              inputError={errorsEmail.email?.message}
              first
            />
            <Button
              disabled={isSubmittingEmail || !isValidEmail}
              type="submit"
              text={isSubmittingEmail ? "" : "Send reset link"}
              size="md"
              className="w-full"
            >
              {isSubmittingEmail ? <Loader width={24} height={24} /> : <></>}
            </Button>
          </Form>
        </>
      )}
      <p className="text-smoke-300 absolute bottom-12 left-1/2 max-w-[400px] -translate-x-1/2 text-center text-[0.875rem]">
        Something went wrong? Contact us on{" "}
        <a
          href="https://t.me/oxytocingroup"
          className="underline-hover text-smoke-100 hover:text-accent relative transition-all"
        >
          Telegram
        </a>
      </p>
    </main>
  );
};
