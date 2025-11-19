import { Button } from "@/components/ui/Button";
import { Input } from "@/components/Form/Input";
import { Link, useNavigate } from "react-router";
import { useEffect } from "react";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import type { SubmitHandler } from "react-hook-form";
import { useForm } from "react-hook-form";
import { useAuth } from "@/hooks/useAuth";
import { api } from "@/api/axiosInstance";
import axios from "axios";
import { Form } from "@/components/Form/Form";
import { Loader } from "@/components/ui/Loader";

const schema = z.object({
  email: z
    .string({ message: "Email is required" })
    .email({ message: "Email is invalid" }),
  username: z
    .string({ message: "Username is required" })
    .min(3, { message: "Username is too short (3 letters min.)" }),
  password: z
    .string({ message: "Password is required" })
    .min(6, { message: "Password is too short (6 letters min.)" }),
});

type FormFields = z.infer<typeof schema>;

export const RegisterPage = () => {
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting, isValid },
  } = useForm<FormFields>({ resolver: zodResolver(schema), mode: "onChange" });

  // const isSubmitting = true;

  const { currentUser } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (currentUser) {
      navigate("/");
    }
  }, [currentUser, navigate]);

  const onSubmit: SubmitHandler<FormFields> = async (data) => {
    try {
      const res = await api.post("/v1/register", data);
      if (res.status == 200) {
        await navigate("/verify");
      }
    } catch (error) {
      if (axios.isAxiosError(error)) {
        if (
          error.response?.data.error.message ===
          "User with this username already exists"
        ) {
          setError("root", { message: "This username already exists" });
        }
        if (
          error.response?.data.error.message ===
          "User with this email already exists"
        ) {
          setError("root", { message: "This email already exists" });
        }
      } else {
        console.log(error);
      }
    }
  };

  return (
    <main className="flex h-full flex-col items-center justify-center gap-6">
      <h1 className="text-[1.25rem]">Welcome to Theca</h1>
      <Form
        onSubmit={handleSubmit(onSubmit)}
        formError={errors.root?.message}
        className="border-smoke-700 flex w-[360px] flex-col items-center justify-center gap-2 rounded-[1.25rem] border-2 p-3"
      >
        <Input
          {...register("email")}
          inputError={errors.email?.message}
          name="email"
          type="email"
          placeholder="Email"
          first
        />
        <Input
          {...register("username")}
          inputError={errors.username?.message}
          name="username"
          type="text"
          placeholder="Username"
        />
        <Input
          {...register("password")}
          inputError={errors.password?.message}
          name="password"
          type="password"
          placeholder="Password"
        />
        <Button
          disabled={isSubmitting || !isValid}
          type="submit"
          text={isSubmitting ? "" : "Sign up"}
          size="md"
          className="w-full leading-6!"
        >
          {isSubmitting ? <Loader width={24} height={24} /> : <></>}
        </Button>
      </Form>
      <p className="text-smoke-300">
        Already have an account?{" "}
        <Link
          to={"/login"}
          className="underline-hover text-smoke-100 hover:text-accent relative transition-all"
        >
          Log in
        </Link>
      </p>
    </main>
  );
};
