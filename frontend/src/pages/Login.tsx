import { useAuth } from "../hooks/useAuth";
import { useForm, SubmitHandler } from "react-hook-form";

type Inputs = {
  email: string;
  password: string;
};

export function Login() {
  const { login } = useAuth();
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<Inputs>();
  const onSubmit: SubmitHandler<Inputs> = async (data) => {
    const result = await login({
      ...data,
    });
    if (!result) {
      console.log("error");
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)}>
      <input
        type="email"
        defaultValue=""
        {...register("email", {
          required: true,
        })}
      />
      {errors.email && <span>This field is required</span>}
      <input type="password" {...register("password", { required: true })} />
      {errors.password && <span>This field is required</span>}
      <input type="submit" />
    </form>
  );
}
