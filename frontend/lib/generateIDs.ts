export function generateTrackID(kind : "audio" | "video" | "screen"){
    return `${kind}-${crypto.randomUUID()}`
}