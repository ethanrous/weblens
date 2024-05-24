import { Button, Text } from "@mantine/core";
const Fourohfour = () => {
    // const nav = useNavigate()
    const nav = null;
    return (
        <div style={{ height: "50vh", justifyContent: "center" }}>
            <Text style={{ padding: 20 }}>Page not found :(</Text>
            <Button onClick={() => nav("/")} color="primary">
                Go Home
            </Button>
        </div>
    );
};
export default Fourohfour;
