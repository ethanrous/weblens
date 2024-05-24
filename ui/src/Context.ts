import { createContext } from "react";
import { UserContextT } from "./types/Types";

export const UserContext = createContext<UserContextT>(null);
