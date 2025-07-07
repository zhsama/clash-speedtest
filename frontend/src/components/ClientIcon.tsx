import React, { useEffect, useState } from 'react'

interface ClientIconProps {
  icon: React.ComponentType<any>
  className?: string
  size?: number
}

export default function ClientIcon({ icon: Icon, className, size }: ClientIconProps) {
  const [isClient, setIsClient] = useState(false)

  useEffect(() => {
    setIsClient(true)
  }, [])

  if (!isClient) {
    return <div className={className} style={{ width: size, height: size }} />
  }

  return <Icon className={className} size={size} />
}