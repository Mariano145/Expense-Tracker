interface HelloProps {
  name?: string;
}

export default function Hello({ name = "World" }: HelloProps) {
  return <h1>Hello, {name}!</h1>;
}
