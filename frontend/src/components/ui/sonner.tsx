import { Toaster as Sonner } from "sonner"

const Toaster = ({ ...props }: React.ComponentProps<typeof Sonner>) => {
  return (
    <Sonner
      className="toaster group"
      toastOptions={{
        classNames: {
          toast:
            "group toast group-[.toaster]:bg-lavender-900 group-[.toaster]:text-lavender-50 group-[.toaster]:border-lavender-500 group-[.toaster]:shadow-lg",
          description: "group-[.toast]:text-lavender-300",
          actionButton:
            "group-[.toast]:bg-lavender-600 group-[.toast]:text-lavender-50 group-[.toast]:hover:bg-lavender-700",
          cancelButton:
            "group-[.toast]:bg-lavender-800 group-[.toast]:text-lavender-200 group-[.toast]:hover:bg-lavender-700",
          closeButton:
            "group-[.toast]:bg-lavender-800 group-[.toast]:text-lavender-200 group-[.toast]:hover:bg-lavender-700",
        },
      }}
      {...props}
    />
  )
}

export { Toaster }
