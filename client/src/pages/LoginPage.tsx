import { Link, useNavigate } from "react-router";
import { Button } from "@/components/ui/Button";
import { Input } from "@/components/Form/Input";
import { useEffect } from "react";
import { type SubmitHandler, useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { Form } from "@/components/Form/Form";
import axios from "axios";
import { useAuth } from "@/hooks/useAuth";
import { Loader } from "@/components/ui/Loader";

const schema = z.object({
  username: z
    .string({ message: "Username is required" })
    .min(3, { message: "Username is too short (3 letters min.)" }),
  password: z
    .string({ message: "Password is required" })
    .min(6, { message: "Password is too short (6 letters min.)" }),
});

type FormFields = z.infer<typeof schema>;

export const LoginPage = () => {
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting, isValid },
  } = useForm<FormFields>({ resolver: zodResolver(schema), mode: "onChange" });
  const { currentUser, handleLogin } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (currentUser) {
      navigate("/");
    }
  }, [currentUser, navigate]);

  const onSubmit: SubmitHandler<FormFields> = async (data) => {
    try {
      await handleLogin(data);
    } catch (error) {
      if (axios.isAxiosError(error)) {
        console.log(error);
        if (error.response?.data.error.message === "Email not verified") {
          navigate("/verify");
        }
        if (
          error.response?.data.error.message === "Invalid username or password"
        ) {
          setError("root", { message: "Invalid username or password" });
        } else {
          setError("root", {
            message: "Something went wrong, please try again.",
          });
        }
      } else {
        setError("root", {
          message: "Something went wrong, please try again.",
        });
        console.log(error);
      }
    }
  };

  return (
    <main className="flex h-full flex-col items-center justify-center gap-6">
      <h1 className="text-[1.25rem]">Welcome back to Theca</h1>
      <Form formError={errors.root?.message} onSubmit={handleSubmit(onSubmit)}>
        <Input
          {...register("username")}
          name="username"
          type="text"
          placeholder="Username"
          inputError={errors.username?.message}
          first
        />
        <Input
          {...register("password")}
          name="password"
          type="password"
          placeholder="Password"
          inputError={errors.password?.message}
        />
        <Button
          disabled={isSubmitting || !isValid}
          type="submit"
          text={isSubmitting ? "" : "Log in"}
          size="md"
          className="w-full"
        >
          {isSubmitting ? <Loader width={24} height={24} /> : <></>}
        </Button>
      </Form>
      <div className="flex flex-col items-center gap-3">
        <p className="text-smoke-300">
          Forgot your password?{" "}
          <Link
            to={"/reset"}
            className="underline-hover text-smoke-100 hover:text-accent relative transition-all"
          >
            Reset
          </Link>
        </p>
        <p className="text-smoke-300">
          Don't have an account yet?{" "}
          <Link
            to={"/register"}
            className="underline-hover text-smoke-100 hover:text-accent relative transition-all"
          >
            Sign up
          </Link>
        </p>
      </div>
    </main>
  );
};
