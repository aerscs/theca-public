import { OTPInput } from "input-otp";
import { Button } from "@/components/ui/Button";
import { Slot } from "@/components/Form/OTPSlot";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm, type SubmitHandler } from "react-hook-form";
import { z } from "zod";
import { useEffect, useState } from "react";
import { api } from "@/api/axiosInstance";
import { useNavigate } from "react-router";
import { cn } from "@/lib/utils";
import { Loader } from "@/components/ui/Loader";
import { Input } from "@/components/Form/Input";

const codeSchema = z.object({
  code: z.string().min(6),
});

const emailSchema = z.object({
  email: z
    .string({ message: "Email is required" })
    .email({ message: "Email is invalid" }),
});

type CodeFormFields = z.infer<typeof codeSchema>;
type EmailFormFields = z.infer<typeof emailSchema>;

export const VerifyPage = () => {
  const [cooldown, setCooldown] = useState(10);
  const [isResending, setIsResending] = useState(false);
  const navigate = useNavigate();

  const {
    register: registerCode,
    handleSubmit: handleSubmitCode,
    formState: { isSubmitting: isSubmittingCode, isValid: isValidCode },
  } = useForm<CodeFormFields>({
    resolver: zodResolver(codeSchema),
    mode: "onChange",
  });

  const {
    register,
    handleSubmit,
    formState: { isSubmitting, errors, isValid },
  } = useForm<EmailFormFields>({
    resolver: zodResolver(emailSchema),
    mode: "onChange",
  });

  const onSubmitCode: SubmitHandler<CodeFormFields> = async (data) => {
    try {
      await api.patch("/v1/verify-email", data);
      navigate("/");
    } catch (error) {
      console.log(error);
    }
  };

  const onSubmitEmail: SubmitHandler<EmailFormFields> = async (data) => {
    await api.post("/v1/send-email-verification-code", data);
    setCooldown(60);
    setIsResending(false);
  };

  useEffect(() => {
    if (cooldown === 0) return;

    const timer = setInterval(() => {
      setCooldown((prev) => prev - 1);
    }, 1000);

    return () => clearInterval(timer);
  }, [cooldown]);

  const handleClick = async () => {
    if (cooldown > 0) return;
    setIsResending(true);
    setCooldown(60);
  };

  return (
    <main className="flex h-full w-[400px] flex-col items-center justify-center gap-6">
      {isResending ? (
        <>
          <div className="flex flex-col items-center gap-3 text-center">
            <h1 className="text-[1.25rem]">Resend verification code</h1>
            <p className="text-smoke-300 text-[0.875rem]">
              Enter your email address to request new verification code.
            </p>
          </div>
          <form
            className="border-smoke-700 bg-smoke-900 flex w-[360px] flex-col items-start justify-center gap-2 rounded-[1.25rem] border-2 p-3 transition-all"
            onSubmit={handleSubmit(onSubmitEmail)}
            {...register("email")}
          >
            <Input
              {...register("email")}
              name="email"
              type="text"
              placeholder="Email"
              inputError={errors.email?.message}
              first
            />
            <Button
              disabled={isSubmitting || !isValid}
              type="submit"
              text={isSubmitting ? "" : "Resend code"}
              size="md"
              className="w-full"
            >
              {isSubmitting ? <Loader width={24} height={24} /> : <></>}
            </Button>
          </form>
        </>
      ) : (
        <>
          <div className="flex flex-col items-center gap-3 text-center">
            <h1 className="text-[1.25rem]">Verify Your Account</h1>
            <p className="text-smoke-300 text-[0.875rem]">
              We’ve sent a verification code to your email. Please enter it
              below to continue.
            </p>
          </div>
          <form
            className="border-smoke-700 bg-smoke-900 flex w-[360px] flex-col items-start justify-center gap-2 rounded-[1.25rem] border-2 p-3 transition-all"
            onSubmit={handleSubmitCode(onSubmitCode)}
            {...registerCode("code")}
          >
            <OTPInput
              name="code"
              autoFocus
              maxLength={6}
              containerClassName="group flex items-center w-full gap-2"
              inputMode="numeric"
              render={({ slots }) =>
                slots.map((slot, idx) => <Slot key={idx} {...slot} />)
              }
            />
            <Button
              disabled={isSubmittingCode || !isValidCode}
              type="submit"
              text={isSubmittingCode ? "" : "Verify"}
              size="md"
              className="w-full"
            >
              {isSubmittingCode ? <Loader width={24} height={24} /> : <></>}
            </Button>
          </form>
          <p className="text-smoke-100 text-[0.875rem] font-medium">
            <button
              disabled={cooldown !== 0}
              onClick={handleClick}
              className={cn(
                "disabled:text-smoke-300 text-smoke-100 hover:text-accent relative transition-all",
                cooldown === 0 && "underline-hover",
              )}
            >
              Resend code
            </button>{" "}
            <span className="text-smoke-100">({cooldown})</span>
          </p>
        </>
      )}
    </main>
  );
};
