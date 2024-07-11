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
    /* "handleSubmit" will validate your inputs before invoking "onSubmit" */
    <form onSubmit={handleSubmit(onSubmit)}>
      {/* register your input into the hook by invoking the "register" function */}
      <input
        type="email"
        defaultValue=""
        {...register("email", {
          required: true,
        })}
      />
      {errors.email && <span>This field is required</span>}

      {/* include validation with required or other standard HTML validation rules */}
      <input type="password" {...register("password", { required: true })} />
      {/* errors will return when field validation fails  */}
      {errors.password && <span>This field is required</span>}

      <input type="submit" />
    </form>
  );
}
