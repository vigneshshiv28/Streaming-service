export interface User{
    id: string;
    name: string;
    email : string;
}

interface Room{
    userId:string | null;
    name:string | null;
    role:string | null;
    roomId:string | null;
    wsURL: string | null
    hostURL:string | null;
    guestURL:string | null;
    audienceURL:string | null;
    createdAt:string | null;
}
