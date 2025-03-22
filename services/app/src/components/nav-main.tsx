"use client"

import { type LucideIcon } from "lucide-react"
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
} from "@/components/ui/sidebar"

export function NavMain({
  items,
}: {
  items: {
    title: string
    url: string
    icon?: LucideIcon
    isActive?: boolean
    items?: {
      title: string
      url: string
    }[]
  }[]
}) {
  return (
    <SidebarGroup>
      <div className="flex items-center justify-center h-16">
        <img src="/logo.svg" width={140} />
      </div>
      <SidebarGroupLabel>Platform</SidebarGroupLabel>
      <SidebarGroupContent>
      <SidebarMenu>
        {items.map((item) => (
            <SidebarMenuItem key={item.title}>
            <a href={item.url}>
              <SidebarMenuButton tooltip={item.title} isActive={item.isActive}>
                {item.icon && <item.icon />}
                <span>{item.title}</span>
              </SidebarMenuButton>
            </a>
          </SidebarMenuItem>
        ))}
      </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  )
}
