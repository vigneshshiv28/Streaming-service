"use client"
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useUserStore } from '@/store/userStore'
import { useRoomStore } from '@/store/roomStore'
import { createRoom, joinRoom } from '@/api/rooms'
import { useRouter } from 'next/navigation'


export default function Dashboard() {
  const [isJoinDisable,setIsJoinDisable] = useState(true) 
  const room = useRoomStore((state)=>state.room)
  const setRoom = useRoomStore((state)=>state.setRoom) 
  const [roomLink, setRoomLink] = useState('')
  const user = useUserStore((state)=>state.user)
  const router = useRouter()

  async function handleCreateRoom(){
    console.log("creating room")
    try{
        const data = await createRoom({userId:user?.id, name:user?.name})
        console.log("room:",data)
        setRoom(data)
        setIsJoinDisable(false)
    } catch(err){
        console.log("error:",err)
    }
  }  

  async function handleJoinRoom(){
    try{
        const data = await joinRoom({userId:user?.id, roomId: room?.roomId, role:"host"})
        console.log("join details:",data)
        if (data.wsUrl && room){
          room.wsURL = data.wsUrl
          router.push(`/room/${room.roomId}`)
        }
           

        
    }catch(err){
        console.log("error:",err)
    }
  }

  async function handleJoinViaLink(){
    if (!roomLink.trim()) {
      console.log("Please enter a room link")
      return
    }
    
    try {
      
      const { roomId, role } = extractRoomDetailsFromLink(roomLink)
      
      if (!roomId) {
        console.log("Invalid room link format")
        return
      }
      
      const data = await joinRoom({userId: user?.id, roomId: roomId, role: role})
      console.log("join via link details:", data)
      
      if (data.wsUrl) {
      
        setRoom({...data, roomId: roomId, wsURL: data.wsUrl})
        router.push(`/room/${roomId}`)
      }
    } catch(err) {
      console.log("error joining via link:", err)
    }
  }


  function extractRoomDetailsFromLink(link:string) {
    try {
      const url = new URL(link)
      

      const pathSegments = url.pathname.split('/')
      const joinIndex = pathSegments.indexOf('join')
      
      if (joinIndex === -1 || joinIndex + 1 >= pathSegments.length) {
        throw new Error('Invalid link format: missing join path or room ID')
      }
      
      const roomId = pathSegments[joinIndex + 1]
      
      
      const role = url.searchParams.get('role') || 'guest'
      
      return { roomId, role }
    } catch (error) {
      console.log('Error parsing room link:', error)
      return { roomId: null, role: null }
    }
  }

  if (!user){
    return (<div>Loading....</div>)
  }
  return (
    <div>
        <h1>Dashboard</h1>
        <div>
        <Button onClick={handleCreateRoom} >
            Create Room
        </Button>
        <Button disabled={isJoinDisable} onClick={handleJoinRoom}>
            Join Room
        </Button>
        </div>

        <div>
          <h2>Join via link</h2>
          <Label htmlFor='link'>Link</Label>
          <Input 
            id="link" 
            type="text" 
            value={roomLink}
            onChange={(e)=>{setRoomLink(e.target.value)}}
            placeholder="Enter room link here..."
          />
          <Button onClick={handleJoinViaLink} disabled={!roomLink.trim()}>
            Join
          </Button>
        </div>
    </div>
  )
}

