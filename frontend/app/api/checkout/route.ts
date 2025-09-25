import { NextRequest, NextResponse } from "next/server";

export async function POST(req: NextRequest) {
  const body = await req.json();
  const planCode: string = body.planCode ?? "";

  // TODO: Backend API ile konuşup gerçek checkout oturumu oluşturun.
  // Şimdilik demo URL döndürüyoruz.
  const checkoutUrl = planCode.includes("annual")
    ? "https://checkout.stripe.com/pay/demo-annual"
    : "https://checkout.stripe.com/pay/demo-monthly";

  return NextResponse.json({ provider: "stripe", checkoutUrl });
}
