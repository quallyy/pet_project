package docs

const openapiSpec = `
openapi: 3.0.3
info:
  title: Cargo Platform API
  description: |
    Sourcing marketplace connecting customers in Turkmenistan with dealers who source products from China, Turkey, and Russia.

    ## Authentication
    - Dealers and Admins login via email + password → POST /auth/login
    - Customers login via Phone + OTP (coming soon) → POST /auth/otp/verify
    - All protected routes require Authorization: Bearer <access_token>
    - Access tokens expire in 15 minutes — use POST /auth/refresh to renew
  version: 0.1.0

servers:
  - url: http://localhost:8080
    description: Local (via gateway)

tags:
  - name: Auth
    description: Login, registration, token management
  - name: Orders
    description: Customer order lifecycle — not built yet
  - name: Bids
    description: Dealer bids on orders — not built yet
  - name: Chat
    description: Customer ↔ dealer messaging — not built yet
  - name: Shipments
    description: Tracking — not built yet
  - name: Admin
    description: Platform administration — not built yet

paths:

  /health:
    get:
      summary: Health check
      tags: [Auth]
      security: []
      responses:
        '200':
          description: Gateway is running

  /auth/login:
    post:
      summary: Login — dealers and admins only
      description: Customers use OTP login (not built yet). Returns a token pair on success.
      tags: [Auth]
      security: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/LoginRequest'
            example:
              email: dealer@example.com
              password: securepassword123
      responses:
        '200':
          description: Login successful
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TokenPair'
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'

  /auth/refresh:
    post:
      summary: Refresh access token
      description: Exchange a valid refresh token for a new token pair. Old refresh token is invalidated immediately.
      tags: [Auth]
      security: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/RefreshRequest'
      responses:
        '200':
          description: New token pair issued
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TokenPair'
        '401':
          $ref: '#/components/responses/Unauthorized'

  /auth/logout:
    post:
      summary: Logout current session
      description: Revokes the session and blacklists the JWT immediately, even before it expires.
      tags: [Auth]
      security:
        - bearerAuth: []
      responses:
        '200':
          description: Logged out
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/MessageResponse'
        '401':
          $ref: '#/components/responses/Unauthorized'

  /auth/logout-all:
    post:
      summary: Logout all sessions
      description: Terminates every active session for this user. Use when phone is stolen or account is compromised.
      tags: [Auth]
      security:
        - bearerAuth: []
      responses:
        '200':
          description: All sessions terminated
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/MessageResponse'
        '401':
          $ref: '#/components/responses/Unauthorized'

  /auth/dealer/register:
    post:
      summary: Complete dealer registration
      description: |
        Dealers cannot self-register. Admin generates an invite link first.
        Dealer uses the token from that link here to set their password and activate account.
        Invite token is single-use and expires after 48 hours.
      tags: [Auth]
      security: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/DealerRegisterRequest'
            example:
              invite_token: "550e8400e29b41d4a716446655440000"
              password: "securepassword123"
      responses:
        '201':
          description: Registered and logged in
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TokenPair'
        '400':
          $ref: '#/components/responses/BadRequest'
        '409':
          description: Email already registered

  /auth/admin/invite:
    post:
      summary: Create dealer invite link — admin only
      description: |
        Generates a single-use invite link for a new dealer.
        Link expires after 48 hours.
      tags: [Auth]
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateInviteRequest'
            example:
              email: newdealer@cargo.tm
              dealer_name: "Ashgabat Trading Co."
      responses:
        '201':
          description: Invite created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/InviteResponse'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '403':
          $ref: '#/components/responses/Forbidden'

  /auth/otp/send:
    post:
      summary: Send OTP — NOT BUILT YET
      description: Sends a 6-digit OTP via SMS. Rate limited to 5 per hour per phone.
      tags: [Auth]
      security: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [phone]
              properties:
                phone:
                  type: string
                  example: "+99361234567"
      responses:
        '200':
          description: OTP sent
        '429':
          description: Rate limit exceeded

  /auth/otp/verify:
    post:
      summary: Verify OTP and login — NOT BUILT YET
      description: Verifies OTP. Creates customer account on first login. Returns token pair.
      tags: [Auth]
      security: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [phone, code]
              properties:
                phone:
                  type: string
                code:
                  type: string
                  minLength: 6
                  maxLength: 6
      responses:
        '200':
          description: Verified
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TokenPair'
        '401':
          description: Invalid or expired OTP

  /orders:
    post:
      summary: Create order — NOT BUILT YET
      description: Customer uploads a photo and description. All dealers are notified. Order expires in 24h if no bids.
      tags: [Orders]
      security:
        - bearerAuth: []
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              type: object
              required: [photo, description]
              properties:
                photo:
                  type: string
                  format: binary
                description:
                  type: string
      responses:
        '201':
          description: Order created
        '401':
          $ref: '#/components/responses/Unauthorized'

    get:
      summary: List my orders — NOT BUILT YET
      description: Returns all orders for the authenticated customer.
      tags: [Orders]
      security:
        - bearerAuth: []
      responses:
        '200':
          description: List of orders

  /orders/{orderId}:
    get:
      summary: Get order details — NOT BUILT YET
      tags: [Orders]
      security:
        - bearerAuth: []
      parameters:
        - $ref: '#/components/parameters/OrderId'
      responses:
        '200':
          description: Order details
        '404':
          $ref: '#/components/responses/NotFound'

  /orders/{orderId}/cancel:
    post:
      summary: Cancel order — NOT BUILT YET
      description: Customer can cancel anytime before accepting a bid.
      tags: [Orders]
      security:
        - bearerAuth: []
      parameters:
        - $ref: '#/components/parameters/OrderId'
      responses:
        '200':
          description: Order cancelled

  /dealer/orders:
    get:
      summary: List open orders — NOT BUILT YET
      description: Returns all pending orders that have not yet reached agreement. All dealers see all orders.
      tags: [Bids]
      security:
        - bearerAuth: []
      responses:
        '200':
          description: Open orders

  /dealer/orders/{orderId}/bid:
    post:
      summary: Submit a bid — NOT BUILT YET
      description: Dealer submits a price, estimated delivery days, and a note to the customer.
      tags: [Bids]
      security:
        - bearerAuth: []
      parameters:
        - $ref: '#/components/parameters/OrderId'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [price, estimated_days, note]
              properties:
                price:
                  type: number
                estimated_days:
                  type: integer
                note:
                  type: string
      responses:
        '201':
          description: Bid submitted

  /orders/{orderId}/bids:
    get:
      summary: Get bids on my order — NOT BUILT YET
      description: Customer sees all bids submitted by dealers for their order.
      tags: [Bids]
      security:
        - bearerAuth: []
      parameters:
        - $ref: '#/components/parameters/OrderId'
      responses:
        '200':
          description: List of bids

  /orders/{orderId}/bids/{bidId}/accept:
    post:
      summary: Accept a bid — NOT BUILT YET
      description: |
        Customer accepts a bid — opens a chat thread with that dealer.
        Customer can accept multiple bids to chat with multiple dealers simultaneously.
      tags: [Bids]
      security:
        - bearerAuth: []
      parameters:
        - $ref: '#/components/parameters/OrderId'
        - name: bidId
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Bid accepted, chat opened

  /orders/{orderId}/chats:
    get:
      summary: List active chats for an order — NOT BUILT YET
      description: Customer sees all open chat threads for their order (one per dealer they accepted).
      tags: [Chat]
      security:
        - bearerAuth: []
      parameters:
        - $ref: '#/components/parameters/OrderId'
      responses:
        '200':
          description: List of chat threads

  /chats/{chatId}/messages:
    get:
      summary: Get chat messages — NOT BUILT YET
      tags: [Chat]
      security:
        - bearerAuth: []
      parameters:
        - $ref: '#/components/parameters/ChatId'
      responses:
        '200':
          description: Message history

    post:
      summary: Send a message — NOT BUILT YET
      tags: [Chat]
      security:
        - bearerAuth: []
      parameters:
        - $ref: '#/components/parameters/ChatId'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [body]
              properties:
                body:
                  type: string
      responses:
        '201':
          description: Message sent

  /chats/{chatId}/agree:
    post:
      summary: Confirm agreement — NOT BUILT YET
      description: |
        Customer confirms agreement with a dealer in this chat.
        All other open chats for this order are closed.
        All other dealers are notified the order is filled.
      tags: [Chat]
      security:
        - bearerAuth: []
      parameters:
        - $ref: '#/components/parameters/ChatId'
      responses:
        '200':
          description: Agreement confirmed, order moves to payment

  /shipments/{orderId}:
    get:
      summary: Get shipment tracking — NOT BUILT YET
      tags: [Shipments]
      security:
        - bearerAuth: []
      parameters:
        - $ref: '#/components/parameters/OrderId'
      responses:
        '200':
          description: Tracking status and history

  /admin/users:
    get:
      summary: List users — NOT BUILT YET
      tags: [Admin]
      security:
        - bearerAuth: []
      responses:
        '200':
          description: All users

  /admin/dealers:
    get:
      summary: List dealers — NOT BUILT YET
      tags: [Admin]
      security:
        - bearerAuth: []
      responses:
        '200':
          description: All dealers

  /admin/orders:
    get:
      summary: List all orders — NOT BUILT YET
      tags: [Admin]
      security:
        - bearerAuth: []
      responses:
        '200':
          description: All orders

components:

  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
      description: RS256 signed JWT. Expires in 15 minutes. Renew via /auth/refresh.

  parameters:
    OrderId:
      name: orderId
      in: path
      required: true
      schema:
        type: string
        format: uuid
    ChatId:
      name: chatId
      in: path
      required: true
      schema:
        type: string
        format: uuid

  schemas:
    LoginRequest:
      type: object
      required: [email, password]
      properties:
        email:
          type: string
          format: email
        password:
          type: string
          minLength: 8

    RefreshRequest:
      type: object
      required: [refresh_token]
      properties:
        refresh_token:
          type: string

    DealerRegisterRequest:
      type: object
      required: [invite_token, password]
      properties:
        invite_token:
          type: string
        password:
          type: string
          minLength: 8

    CreateInviteRequest:
      type: object
      required: [email, dealer_name]
      properties:
        email:
          type: string
          format: email
        dealer_name:
          type: string

    TokenPair:
      type: object
      properties:
        access_token:
          type: string
          description: JWT — valid for 15 minutes
        refresh_token:
          type: string
          description: Opaque — valid for 30 days, rotated on use
        expires_in:
          type: integer
          example: 900

    InviteResponse:
      type: object
      properties:
        invite_url:
          type: string
          example: "https://cargo.tm/dealer/register?token=abc123"
        expires_in:
          type: string
          example: "48h0m0s"

    MessageResponse:
      type: object
      properties:
        message:
          type: string

    ErrorResponse:
      type: object
      properties:
        error:
          type: string

  responses:
    BadRequest:
      description: Invalid request
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'
    Unauthorized:
      description: Missing or invalid token
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'
    Forbidden:
      description: Insufficient permissions
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'
    NotFound:
      description: Not found
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'
`